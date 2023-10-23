package withdrawals

import (
	"context"

	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/logger"
)

type Storage interface {
	GetAll(ctx context.Context, userID int64) ([]Withdrowal, error)
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

func (u *usecase) GetAll(ctx context.Context, userID int64) ([]Withdrowal, error) {
	return u.storage.GetAll(ctx, userID)
}
