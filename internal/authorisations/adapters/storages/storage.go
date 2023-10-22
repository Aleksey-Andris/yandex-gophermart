package storages

import (
	"context"
	"errors"

	"github.com/Aleksey-Andris/yandex-gophermart/internal/authorisations"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/db"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/logger"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
	query := "INSERT INTO ygm_user (login, password_hash) VALUES($1, $2) RETURNING id;"
	row := s.db.Pool.QueryRow(ctx, query, auth.Login, auth.Password)
	err := row.Scan(&auth.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			err = db.ErrConflict
		}
	}
	return auth, err
}

func (s *storage) Login(ctx context.Context, auth *authorisations.Auth) (*authorisations.Auth, error) {
	query := "SELECT id, login, password_hash FROM ygm_user WHERE login=$1 AND password_hash=$2;"
	row := s.db.Pool.QueryRow(ctx, query, auth.Login, auth.Password)
	err := row.Scan(&auth.ID, &auth.Login, &auth.Password)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = db.ErrNoRows
		}
	}
	return auth, err
}
