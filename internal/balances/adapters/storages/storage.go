package storages

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/Aleksey-Andris/yandex-gophermart/internal/authorisations"
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

func (s *storage) Spend(ctx context.Context, oper *balances.Operation) (*balances.Operation, error) {
	var factUserId int64
	query := "SELECT id FROM ygm_user WHERE id = $1;"
	row := s.db.Pool.QueryRow(ctx, query, authorisations.GetUserID(ctx))
	err := row.Scan(&factUserId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, db.ErrUserNotExist
		}
		return nil, err
	}

	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	ordNum, err := strconv.ParseInt(string(oper.Ord), 10, 64)
	if err != nil {
		return nil, err
	}
	var ordID int64
	query = "INSERT INTO ygm_order (user_id, status_id, num, order_date)" +
		" VALUES($1, (SELECT s.id FROM ygm_order_status s WHERE s.ident=$2), $3, $4) RETURNING id;"
	row = tx.QueryRow(ctx, query, factUserId, "NEW", ordNum, time.Now())
	err = row.Scan(&ordID)
	if err != nil {
		return nil, err
	}

	query = "INSERT INTO ygm_balls_operation (order_id, amount) VALUES($1, $2);"
	_, err = tx.Exec(ctx, query, ordID, -oper.Amount)
	if err != nil {
		return nil, err
	}

	var sum float64
	query = "SELECT COALESCE(SUM(opp.amount), 0) AS current FROM ygm_balls_operation opp" +
		" LEFT JOIN ygm_order ordp ON ordp.id = opp.order_id" +
		" WHERE ordp.user_id = $1;"
	row = tx.QueryRow(ctx, query, factUserId)
	err = row.Scan(&sum)
	if err != nil {
		return nil, err
	}
	if sum < 0 {
		oper.Result = balances.ResultNotEnough
		return oper, err
	}
	oper.Result = balances.ResultOK
	err = tx.Commit(ctx)
	return oper, err
}
