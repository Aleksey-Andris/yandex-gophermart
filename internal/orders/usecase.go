package orders

import (
	"context"

	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/logger"
)

type Storage interface {
	AddOne(ctx context.Context, auth *Order) (*Order, error)
	GetAll(ctx context.Context, auth *Order) (*Order, error)
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

func (u *usecase) AddOne(ctx context.Context, auth *Order) (*Order, error) {
	return u.storage.AddOne(ctx, auth)
}

func (u *usecase) GetAll(ctx context.Context, auth *Order) (*Order, error) {
	return u.storage.GetAll(ctx, auth)
}
