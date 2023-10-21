package authorisations

import (
	"context"

	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/logger"
)

type Storage interface {
	Register(ctx context.Context, auth *Auth) (*Auth, error)
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

func (u *usecase) Register(ctx context.Context, auth *Auth) (*Auth, error) {
	return u.storage.Register(ctx, auth)
}
