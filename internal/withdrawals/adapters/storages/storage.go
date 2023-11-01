package storages

import (
	"context"
	"fmt"

	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/db"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/logger"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/withdrawals"
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
	operations := make([]withdrawals.Withdrowal, 0)
	query := "SELECT ord.num, @ op.amount, ord.order_date" +
		" FROM ygm_balls_operation op" +
		" INNER JOIN ygm_order ord ON ord.id = op.order_id" +
		" WHERE  ord.user_id = $1 AND op.amount < 0 ORDER BY ord.order_date;"
	rows, err := s.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get withdrawals from DB: %w", err)
	}
	for rows.Next() {
		oper := withdrawals.Withdrowal{}
		if err := rows.Scan(&oper.Ord, &oper.Amount, &oper.Data); err != nil {
			return nil, fmt.Errorf("failed to scan withdrawals from rows: %w", err)
		}
		operations = append(operations, oper)
	}
	return operations, nil
}
