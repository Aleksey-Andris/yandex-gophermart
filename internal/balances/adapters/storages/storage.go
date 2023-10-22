package storages

import (
	"context"
	"errors"

	"github.com/Aleksey-Andris/yandex-gophermart/internal/balances"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/db"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/logger"
	"github.com/jackc/pgx/v5"
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

func (s *storage) Get(ctx context.Context, userID int64) (*balances.Balance, error) {
	var factUserId int64
	query := "SELECT id FROM ygm_user WHERE id = $1;"
	row := s.db.Pool.QueryRow(ctx, query, userID)
	err := row.Scan(&factUserId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, db.ErrUserNotExist
		}
		return nil, err
	}

	query = "SELECT" +

		" (SELECT COALESCE(SUM(opp.amount), 0) AS current FROM ygm_balls_operation opp" +
		" LEFT JOIN ygm_order ordp ON ordp.id = opp.order_id" +
		" WHERE ordp.user_id = $1)," +

		" (SELECT @ COALESCE(SUM(opm.amount), 0) AS withdrawn  FROM ygm_balls_operation opm" +
		" LEFT JOIN ygm_order ordm ON ordm.id = opm.order_id" +
		" WHERE ordm.user_id = $2 AND opm.amount < 0);"

	var balanse balances.Balance
	row = s.db.Pool.QueryRow(ctx, query, userID, userID)
	err = row.Scan(&balanse.Current, &balanse.Withdrawn)
	return &balanse, err
}
