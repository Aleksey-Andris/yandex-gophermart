package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/Aleksey-Andris/yandex-gophermart/internal/authorisations"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/balances"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/logger"
	"github.com/go-chi/chi"
)

const (
	сontentType         = "Content-Type"
	сontentTypeAppJSON  = "application/json"
	сontentTypeAppXGZIP = "application/x-gzip"
)

type Usecase interface {
	Get(ctx context.Context, userID int64) (*balances.Balance, error)
	Spend(ctx context.Context, oper *balances.Operation) (*balances.Operation, error)
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
	router.Get("/", c.get)
	router.Post("/withdraw", c.spend)
	return router
}

func (c *controller) get(res http.ResponseWriter, req *http.Request) {
	balanses, err := c.usecase.Get(req.Context(), authorisations.GetUserID(req.Context()))
	if err != nil {
		c.logger.Errorf("Balanses: failed to get orders, err value: %s. Session ID: %s", err, c.logger.GetSesionID(req.Context()))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(&balanses)
	if err != nil {
		c.logger.Errorf("Balanses: failed to marshal body, err value: %s. Session ID: %s", err, c.logger.GetSesionID(req.Context()))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	res.Header().Set(сontentType, сontentTypeAppJSON)
	res.WriteHeader(http.StatusOK)
	res.Write(response)
}

func (c *controller) spend(res http.ResponseWriter, req *http.Request) {
	ct := strings.Split(req.Header.Get(сontentType), ";")[0]
	if !(ct == сontentTypeAppJSON || ct == сontentTypeAppXGZIP) {
		c.logger.Errorf("Spend: invalid content type: %s. Session ID: %s", ct, c.logger.GetSesionID(req.Context()))
		http.Error(res, "Invalid Content-Type", http.StatusBadRequest)
		return
	}

	var request balances.Operation
	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		c.logger.Errorf("Spend: failed decoding body, err value: %s. Session ID: %s", err, c.logger.GetSesionID(req.Context()))
		http.Error(res, "Invalid body", http.StatusBadRequest)
		return
	}

	num, err := strconv.ParseInt(string(request.Ord), 10, 64)
	if err != nil || !balances.ValidLoon(int(num)) {
		c.logger.Errorf("Spend: invalid nums format, err value: %s, num: %s. Session ID: %s", err, string(request.Ord), c.logger.GetSesionID(req.Context()))
		http.Error(res, "Invalid nums format", http.StatusUnprocessableEntity)
		return
	}

	operation, err := c.usecase.Spend(req.Context(), &request)
	if err != nil {
		c.logger.Errorf("Spend: failed to spend balls, err value: %s. Session ID: %s", err, c.logger.GetSesionID(req.Context()))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	if operation.Result == balances.ResultNotEnough {
		res.WriteHeader(http.StatusPaymentRequired)
		return
	}
	res.WriteHeader(http.StatusOK)
}
