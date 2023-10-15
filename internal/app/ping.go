package app

import (
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/logger"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/postgres"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/pings"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/pings/adapters/db"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/pings/adapters/http/controllers"
	"github.com/go-chi/chi"
)

func initPing(l *logger.Logger, pg *postgres.Postgres) *chi.Mux {
	pingStorage := db.New(l, pg)
	pingUsecase := pings.New(l, pingStorage)
	pingController := controllers.New(l, pingUsecase)
	return pingController.Init()
}