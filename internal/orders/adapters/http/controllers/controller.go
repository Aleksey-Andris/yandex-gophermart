package controllers

import (
	"context"
	"encoding/json"
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
	AddOne(ctx context.Context, ordrs *orders.Order) (*orders.Order, error)
	GetAll(ctx context.Context, userID int64) ([]orders.Order, error)
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
	router.Get("/", c.getAll)
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
	
	num := string(body)
	numInt, err := strconv.ParseInt(string(body), 10, 64)
	if err != nil || !orders.ValidLoon(int(numInt)) {
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

func (c *controller) getAll(res http.ResponseWriter, req *http.Request) {
	orders, err := c.usecase.GetAll(req.Context(), authorisations.GetUserID(req.Context()))
	if err != nil {
		c.logger.Errorf("Orders: failed to get orders, err value: %s. Session ID: %s", err, c.logger.GetSesionID(req.Context()))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(orders) == 0 {
		res.WriteHeader(http.StatusNoContent)
		return
	}

	response, err := json.Marshal(&orders)
	if err != nil {
		c.logger.Errorf("Orders: failed to marshal body, err value: %s. Session ID: %s", err, c.logger.GetSesionID(req.Context()))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	res.Header().Set(сontentType, сontentTypeAppJSON)
	res.WriteHeader(http.StatusOK)
	res.Write(response)

}
