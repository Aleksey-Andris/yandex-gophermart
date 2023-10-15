package pings

import (
	"context"

	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/logger"
)

type Storage interface {
	Ping(ctx context.Context) error
}

type usecase struct {
	logger  *logger.Logger
	storage Storage
}

func New(logger *logger.Logger, storage Storage) *usecase {
	return &usecase{
		logger:  logger,
		storage: storage,
	}
}

func (u *usecase) Ping(ctx context.Context) error {
	return u.storage.Ping(ctx)
}