package app

import (
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/db"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/logger"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/balances"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/balances/adapters/http/controllers"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/balances/adapters/storages"
	"github.com/go-chi/chi"
)

func initBalance(l *logger.Logger, db *db.Postgres) *chi.Mux {
	storage := storages.New(l, db)
	usecase := balances.New(l, storage)
	controller := controllers.New(l, usecase)
	return controller.Init()
}