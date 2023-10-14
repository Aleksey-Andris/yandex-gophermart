package app

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Aleksey-Andris/yandex-gophermart/config"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/ping/adapters/http/pingcontroller"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/ping/adapters/pingdb"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/ping/pingusecase"
	"github.com/Aleksey-Andris/yandex-gophermart/pkg/httpserver"
	"github.com/Aleksey-Andris/yandex-gophermart/pkg/logger"
	"github.com/Aleksey-Andris/yandex-gophermart/pkg/postgres"
	"github.com/go-chi/chi"
)

func Run(cfg *config.Config) {
	l, err := logger.New(cfg.Log.Environment)
	if err != nil {
		log.Fatalf("logger create error: %s", err)
	}
	defer l.Sync()
	l.Info("config application",
		"Run_Addres:", cfg.HTTP.RunAddres,
		"Accrual_Address:", cfg.HTTP.AccrualAddress,
		"Environment:", cfg.Log.Environment,
		"DB_PoolMax:", cfg.PG.PoolMax,
		"DB_URI:", cfg.PG.URI,
	)

	pg, err := postgres.New(cfg.PG.URI, postgres.MaxPoolSize(cfg.PG.PoolMax))
	if err != nil {
		log.Fatalf("postgres create error: %s", err)
	}

	pingStorage := pingdb.New(l, pg)
	pingUsecase := pingusecase.New(l, pingStorage)
	pingController := pingcontroller.New(l, pingUsecase)

	chiRouter := chi.NewRouter()
	chiRouter.Mount("/", pingController.Init())

	httpServer := httpserver.NewAndStart(chiRouter, httpserver.Address(cfg.HTTP.RunAddres))

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		l.Info("shutting down - signal: %s" + s.String())
	case err = <-httpServer.Notify():
		l.Error("error starting server", err)
	}
	err = httpServer.Shutdown()
	if err != nil {
		l.Error("error shutdowing server", err)
	}
}
