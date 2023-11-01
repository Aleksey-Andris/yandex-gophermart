package app

import (
	"github.com/Aleksey-Andris/yandex-gophermart/internal/pings/adapters/storages"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/db"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/logger"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/pings"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/pings/adapters/http/controllers"
	"github.com/go-chi/chi"
)

func initPing(l *logger.Logger, db *db.Postgres) *chi.Mux {
	storage := storages.New(l, db)
	usecase := pings.New(l, storage)
	сontroller := controllers.New(l, usecase)
	return сontroller.Init()
}