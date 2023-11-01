package app

import (
	"github.com/Aleksey-Andris/yandex-gophermart/internal/authorisations/adapters/storages"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/db"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/logger"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/authorisations"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/authorisations/adapters/http/controllers"
	"github.com/go-chi/chi"
)

func initAuth(l *logger.Logger, db *db.Postgres) *chi.Mux {
	storage := storages.New(l, db)
	usecase := authorisations.New(l, storage)
	controller := controllers.New(l, usecase)
	return controller.Init()
}