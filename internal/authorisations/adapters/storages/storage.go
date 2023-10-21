package storages

import (
	"context"
	"errors"
	"fmt"

	"github.com/Aleksey-Andris/yandex-gophermart/internal/authorisations"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/db"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/logger"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	userTable = "ygm_user"
	userLogin = "login"
	userPass  = "password_hash"
)

type storage struct {
	logger *logger.Logger
	db     *db.Postgres
}

func New(logger *logger.Logger, db *db.Postgres) *storage {
	return &storage{
		logger: logger,
		db:     db,
	}
}

func (s *storage) Register(ctx context.Context, auth *authorisations.Auth) (*authorisations.Auth, error) {
	query := fmt.Sprintf("INSERT INTO %s (%s, %s) VALUES($1, $2) RETURNING id;", userTable, userLogin, userPass)
	row := s.db.Pool.QueryRow(ctx, query, auth.Login, auth.Password)
    err := row.Scan(auth.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			err = db.ErrConflict
		}
	}
	return auth, err
}
 