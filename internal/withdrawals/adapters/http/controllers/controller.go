package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Aleksey-Andris/yandex-gophermart/internal/authorisations"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/db"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/logger"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/withdrawals"
	"github.com/go-chi/chi"
)

const (
	сontentType          = "Content-Type"
	сontentTypeAppJSON   = "application/json"
)

type Usecase interface {
	GetAll(ctx context.Context, userID int64) ([]withdrawals.Withdrowal, error)
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
	router.Get("/", c.getAll)
	return router
}

func (c *controller) getAll(res http.ResponseWriter, req *http.Request) {
	operations, err := c.usecase.GetAll(req.Context(), authorisations.GetUserID(req.Context()))
	if err != nil {
		if errors.Is(err, db.ErrUserNotExist) {
			res.WriteHeader(http.StatusUnauthorized)
			return
		}
		c.logger.Errorf("Withdrawals: failed to get withdrawals, err value: %s. Session ID: %s", err, c.logger.GetSesionID(req.Context()))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(operations) == 0 {
		res.WriteHeader(http.StatusNoContent)
		return
	}

	response, err := json.Marshal(&operations)
	if err != nil {
		c.logger.Errorf("Withdrawals: failed to marshal body, err value: %s. Session ID: %s", err, c.logger.GetSesionID(req.Context()))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	res.Header().Set(сontentType, сontentTypeAppJSON)
	res.WriteHeader(http.StatusOK)
	res.Write(response)

}
