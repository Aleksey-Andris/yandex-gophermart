package storages

import (
	"context"
	"errors"

	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/db"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/logger"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/withdrawals"
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

func (s *storage) GetAll(ctx context.Context, userID int64) ([]withdrawals.Withdrowal, error) {
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

	operations := make([]withdrawals.Withdrowal, 0)
	query = "SELECT ord.num, @ op.amount, ord.order_date" +
		" FROM ygm_balls_operation op" +
		" INNER JOIN ygm_order ord ON ord.id = op.order_id" +
		" WHERE  ord.user_id = $1 AND op.amount < 0 ORDER BY ord.order_date;"
	rows, err := s.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		oper := withdrawals.Withdrowal{}
		if err := rows.Scan(&oper.Ord, &oper.Amount, &oper.Data); err != nil {
			return nil, err
		}
		operations = append(operations, oper)
	}
	return operations, nil
}
