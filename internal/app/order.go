package app

import (
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/db"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/logger"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/orders"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/orders/adapters/http/controllers"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/orders/adapters/storages"
	"github.com/go-chi/chi"
)

func initOrder(l *logger.Logger, db *db.Postgres) *chi.Mux {
	storage := storages.New(l, db)
	usecase := orders.New(l, storage)
	controller := controllers.New(l, usecase)
	return controller.Init()
}
