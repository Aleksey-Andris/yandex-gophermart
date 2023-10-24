package app

import (
	"github.com/Aleksey-Andris/yandex-gophermart/config"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/db"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/logger"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/orders"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/orders/adapters/http/clients"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/orders/adapters/http/controllers"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/orders/adapters/storages"
	"github.com/go-chi/chi"
	"github.com/go-resty/resty/v2"
)

func initOrder(l *logger.Logger, db *db.Postgres, cfg *config.Config) *chi.Mux {
	storage := storages.New(l, db)
	usecase := orders.New(l, storage)
	controller := controllers.New(l, usecase)
	clients.New(l, cfg.HTTP.AccrualAddress,  resty.New(), usecase)
	return controller.Init()
}
