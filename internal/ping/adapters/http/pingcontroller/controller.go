package pingcontroller

import (
	"net/http"

	"github.com/Aleksey-Andris/yandex-gophermart/internal"
	"github.com/go-chi/chi"
)

type controller struct {
	logger  internal.Logger
	usecase internal.PingUsecase
}

func New(logger internal.Logger, usecase internal.PingUsecase) *controller {
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
		c.logger.Error("wrong ping storage", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusOK)
}
