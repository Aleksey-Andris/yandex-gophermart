package app

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Aleksey-Andris/yandex-gophermart/config"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/compression"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/db"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/httpserver"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/logger"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func Run(cfg *config.Config) {
	l, err := logger.New(cfg.LogEnvironment)
	if err != nil {
		log.Fatalf("Logger: failed creating: %s", err)
	}
	defer l.Sync()
	l.Infow("Config application:",
		"Run_Addres", cfg.RunAddres,
		"Accrual_Address", cfg.AccrualAddress,
		"Environment", cfg.LogEnvironment,
		"DB_PoolMax", cfg.DBPoolMax,
		"DB_URI", cfg.DBURI,
	)
	startMigrations(l, cfg)

	pg, err := db.NewPostgres(cfg.DBURI, db.MaxPoolSize(cfg.DBPoolMax))
	if err != nil {
		l.Fatalf("Storage: creating error: %s", err)
	}
	defer pg.Close()

	mux := initRouter(l, pg, cfg)
	httpServer := httpserver.NewAndStart(mux, httpserver.Address(cfg.RunAddres))

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	select {
	case s := <-interrupt:
		l.Infof("Shutdown: signal received: %s", s.String())
	case err = <-httpServer.Notify():
		l.Fatalf("Server: failed starting: %s", err)
	}
	err = httpServer.Shutdown()
	if err != nil {
		l.Fatalf("Server: failed shutdowing: %s", err)
	}
}

func initRouter(l *logger.Logger, pg *db.Postgres, cfg *config.Config) *chi.Mux {
	router := chi.NewRouter()
	router.Use(compression.Decompress)
	router.Use(middleware.Recoverer)
	router.Use(l.WithLogging)
	router.Use(middleware.Compress(5, "application/json", "text/html"))
	router.Mount("/", initPing(l, pg))
	router.Mount("/api/user/", initAuth(l, pg))
	router.Mount("/api/user/orders", initOrder(l, pg, cfg))
	router.Mount("/api/user/balance", initBalance(l, pg))
	router.Mount("/api/user/withdrawals", initWithdrawal(l, pg))
	return router
}
