package app

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Aleksey-Andris/yandex-gophermart/config"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/httpserver"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/logger"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/postgres"
	"github.com/go-chi/chi"
)

func Run(cfg *config.Config) {
	l, err := logger.New(cfg.Log.Environment)
	if err != nil {
		log.Fatalf("logger create error: %s", err)
	}
	defer l.Sync()
	l.Infow("Config application:",
		"Run_Addres", cfg.HTTP.RunAddres,
		"Accrual_Address", cfg.HTTP.AccrualAddress,
		"Environment", cfg.Log.Environment,
		"DB_PoolMax", cfg.PG.PoolMax,
		"DB_URI", cfg.PG.URI,
	)
	startMigrations(l, cfg)

	pg, err := postgres.New(cfg.PG.URI, postgres.MaxPoolSize(cfg.PG.PoolMax))
	if err != nil {
		l.Fatalf("postgres creating error: %s", err)
	}
	defer pg.Close()

	mux := initMux(l, pg)
	httpServer := httpserver.NewAndStart(mux, httpserver.Address(cfg.HTTP.RunAddres))

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	select {
	case s := <-interrupt:
		l.Infof("shutting down signal: ", s.String())
	case err = <-httpServer.Notify():
		l.Fatal("error starting server: %s", err)
	}
	err = httpServer.Shutdown()
	if err != nil {
		l.Fatalf("error shutdowing server: %s", err)
	}
}

func initMux(l *logger.Logger, pg *postgres.Postgres) *chi.Mux {
	chiRouter := chi.NewRouter()
	chiRouter.Mount("/", initPing(l, pg))
	return chiRouter
}
