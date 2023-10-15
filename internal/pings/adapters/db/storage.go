package db

import (
	"context"

	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/logger"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/postgres"
)

type storage struct {
	logger *logger.Logger
	db     *postgres.Postgres
}

func New(logger *logger.Logger, db *postgres.Postgres) *storage {
	return &storage{
		logger: logger,
		db:     db,
	}
}

func (s *storage) Ping(ctx context.Context) error {
	return s.db.Pool.Ping(ctx)
}
