package controllers

import (
	"context"
	"net/http"

	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/logger"
	"github.com/go-chi/chi"
)

type Usecase interface {
	Ping(ctx context.Context) error
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
	router.Get("/ping", c.ping)
	return router
}

func (c *controller) ping(res http.ResponseWriter, req *http.Request) {
	err := c.usecase.Ping(req.Context())
	if err != nil {
		c.logger.Errorf("error pinging storage: %s", err)
		res.WriteHeader(http.StatusInternalServerError)
		return 
	}
	res.WriteHeader(http.StatusOK)
}
