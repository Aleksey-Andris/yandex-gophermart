package storages

import (
	"context"
	"fmt"
	"time"

	"github.com/Aleksey-Andris/yandex-gophermart/internal/authorisations"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/balances"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/db"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/logger"
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
	query := "SELECT" +

		" (SELECT COALESCE(SUM(opp.amount), 0) AS current FROM ygm_balls_operation opp" +
		" LEFT JOIN ygm_order ordp ON ordp.id = opp.order_id" +
		" WHERE ordp.user_id = $1)," +

		" (SELECT @ COALESCE(SUM(opm.amount), 0) AS withdrawn  FROM ygm_balls_operation opm" +
		" LEFT JOIN ygm_order ordm ON ordm.id = opm.order_id" +
		" WHERE ordm.user_id = $2 AND opm.amount < 0);"

	var balanse balances.Balance
	row := s.db.Pool.QueryRow(ctx, query, userID, userID)
	err := row.Scan(&balanse.Current, &balanse.Withdrawn)
	if err != nil {
		err = fmt.Errorf("failed to get balanse from DB: %w", err)
	}
	return &balanse, err
}

func (s *storage) Spend(ctx context.Context, oper *balances.Operation) (*balances.Operation, error) {
	userID := authorisations.GetUserID(ctx)

	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := "SELECT id FROM ygm_user WHERE id=$1 FOR UPDATE;"
	row := tx.QueryRow(ctx, query, userID)
	err = row.Scan(&userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user from DB: %w", err)
	}

	var ordID int64
	query = "INSERT INTO ygm_order (user_id, status_id, num, order_date)" +
		" VALUES($1, (SELECT s.id FROM ygm_order_status s WHERE s.ident=$2), $3, $4) RETURNING id;"
	row = tx.QueryRow(ctx, query, userID, "NEW", oper.Ord, time.Now())
	err = row.Scan(&ordID)
	if err != nil {
		return nil, fmt.Errorf("failed to insert order in DB: %w", err)
	}

	query = "INSERT INTO ygm_balls_operation (order_id, amount) VALUES($1, $2);"
	_, err = tx.Exec(ctx, query, ordID, -oper.Amount)
	if err != nil {
		return nil, fmt.Errorf("failed to insert operetion in DB: %w", err)
	}

	var sum float64
	query = "SELECT COALESCE(SUM(opp.amount), 0) AS current FROM ygm_balls_operation opp" +
		" LEFT JOIN ygm_order ordp ON ordp.id = opp.order_id" +
		" WHERE ordp.user_id = $1;"
	row = tx.QueryRow(ctx, query, userID)
	err = row.Scan(&sum)
	if err != nil {
		return nil, fmt.Errorf("failed to check balls sum: %w", err)
	}
	if sum < 0 {
		oper.Result = balances.ResultNotEnough
		return oper, err
	}
	oper.Result = balances.ResultOK
	err = tx.Commit(ctx)
	return oper, err
}
