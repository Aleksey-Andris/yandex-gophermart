package balances

import (
	"context"

	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/logger"
)

type Storage interface {
	Get(ctx context.Context, userID int64) (*Balance, error)
	Spend(ctx context.Context, oper *Operation) (*Operation, error)
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

func (u *usecase) Get(ctx context.Context, userID int64) (*Balance, error) {
	return u.storage.Get(ctx, userID)
}
func (u *usecase) Spend(ctx context.Context, oper *Operation) (*Operation, error) {
	return u.storage.Spend(ctx, oper)
}
