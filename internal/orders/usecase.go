package orders

import (
	"context"

	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/logger"
)

type Storage interface {
	AddOne(ctx context.Context, auth *Order) (*Order, error)
	GetAll(ctx context.Context, userID int64) ([]Order, error)
	GetAllUactual(ctx context.Context) ([]Order, error)
	Update(ctx context.Context, ordrs []Order) error
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

func (u *usecase) AddOne(ctx context.Context, order *Order) (*Order, error) {
	return u.storage.AddOne(ctx, order)
}

func (u *usecase) GetAll(ctx context.Context, userID int64) ([]Order, error) {
	return u.storage.GetAll(ctx, userID)
}

func (u *usecase) GetAllUactual(ctx context.Context) ([]Order, error) {
	return u.storage.GetAllUactual(ctx)
}

func (u *usecase) Update(ctx context.Context, ordrs []Order) error {
	return u.storage.Update(ctx, ordrs)
}
