package pingusecase

import (
	"context"

	"github.com/Aleksey-Andris/yandex-gophermart/internal"
)

type usecase struct {
	logger  internal.Logger
	storage internal.PingStorage
}

func New(logger internal.Logger, storage internal.PingStorage) *usecase {
	return &usecase{
		logger:  logger,
		storage: storage,
	}
}

func (u *usecase) Ping(ctx context.Context) error {
	return u.storage.Ping(ctx)
}
