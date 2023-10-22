package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Aleksey-Andris/yandex-gophermart/internal/authorisations"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/balances"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/db"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/logger"
	"github.com/go-chi/chi"
)

const (
	сontentType          = "Content-Type"
	сontentTypeAppJSON   = "application/json"
)

type Usecase interface {
	Get(ctx context.Context, userID int64) (*balances.Balance, error)
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
	return router
}

func (c *controller) get(res http.ResponseWriter, req *http.Request) {
	balanses, err := c.usecase.Get(req.Context(), authorisations.GetUserID(req.Context()))
	if err != nil {
		if errors.Is(err, db.ErrUserNotExist) {
			res.WriteHeader(http.StatusUnauthorized)
			return
		}
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