package controllers

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Aleksey-Andris/yandex-gophermart/internal/authorisations"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/db"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/logger"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/orders"
	"github.com/go-chi/chi"
)

const (
	сontentType          = "Content-Type"
	сontentTypeTextPlain = "text/plain"
	сontentTypeAppJSON   = "application/json"
	сontentTypeAppXGZIP  = "application/x-gzip"
)

type Usecase interface {
	AddOne(ctx context.Context, auth *orders.Order) (*orders.Order, error)
	GetAll(ctx context.Context, auth *orders.Order) (*orders.Order, error)
}

type controller struct {
	logger  *logger.Logger
	usecase Usecase
}

func New(logger *logger.Logger, usecase Usecase) *controller {
	return &controller{
		logger:  logger,
		usecase: usecase,
	}
}

func (c *controller) Init() *chi.Mux {
	router := chi.NewRouter()
	router.Use(authorisations.UserIdentity)
	router.Post("/", c.addOne)
	//router.Get("/", c.getAll)
	return router
}

func (c *controller) addOne(res http.ResponseWriter, req *http.Request) {
	ct := strings.Split(req.Header.Get(сontentType), ";")[0]
	if !(ct == сontentTypeTextPlain || ct == сontentTypeAppXGZIP) {
		c.logger.Errorf("Orders: invalid content type: %s. Session ID: %s", ct, c.logger.GetSesionID(req.Context()))
		http.Error(res, "Invalid Content-Type", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		c.logger.Errorf("Orders: failed reading bodyy, err value:: %s. Session ID: %s", err, c.logger.GetSesionID(req.Context()))
		http.Error(res, "Invalid body", http.StatusBadRequest)
		return
	}

	num, err := strconv.ParseInt(string(body), 10, 64)
	if err != nil || !validLoon(int(num)) {
		c.logger.Errorf("Orders: invalid nums format, err value: %s, num: %s. Session ID: %s", err, string(body), c.logger.GetSesionID(req.Context()))
		http.Error(res, "Invalid nums format", http.StatusUnprocessableEntity)
		return
	}

	order := &orders.Order{
		UserID:      authorisations.GetUserID(req.Context()),
		StatusIdent: orders.StatusNew,
		Num:         num,
		Date:        time.Now(),
	}
	_, err = c.usecase.AddOne(req.Context(), order)
	if err != nil {
		if errors.Is(err, db.ErrRowExist) {
			res.WriteHeader(http.StatusOK)
			return
		}
		if errors.Is(err, db.ErrConflict) {
			res.WriteHeader(http.StatusConflict)
			return
		}
		c.logger.Errorf("Orders: failed to add order, err value: %s. Session ID: %s", err, c.logger.GetSesionID(req.Context()))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusAccepted)
}

func validLoon(number int) bool {
	return (number%10+checksum(number/10))%10 == 0
}

func checksum(number int) int {
	var luhn int
	for i := 0; number > 0; i++ {
		cur := number % 10
		if i%2 == 0 {
			cur = cur * 2
			if cur > 9 {
				cur = cur%10 + cur/10
			}
		}
		luhn += cur
		number = number / 10
	}
	return luhn % 10
}
