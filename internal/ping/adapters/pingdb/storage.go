package pingdb

import (
	"context"

	"github.com/Aleksey-Andris/yandex-gophermart/internal"
	"github.com/Aleksey-Andris/yandex-gophermart/pkg/postgres"
)

type storage struct {
	logger internal.Logger
	db     *postgres.Postgres
}

func New(logger internal.Logger, db *postgres.Postgres) *storage {
	return &storage{
		logger: logger,
		db:     db,
	}
}

func (s *storage) Ping(ctx context.Context) error {
	return s.db.Pool.Ping(ctx)
}
