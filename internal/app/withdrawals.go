package app

import (
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/db"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/logger"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/withdrawals"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/withdrawals/adapters/http/controllers"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/withdrawals/adapters/storages"
	"github.com/go-chi/chi"
)

func initWithdrawal(l *logger.Logger, db *db.Postgres) *chi.Mux {
	storage := storages.New(l, db)
	usecase := withdrawals.New(l, storage)
	controller := controllers.New(l, usecase)
	return controller.Init()
}