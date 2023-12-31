package storages

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/db"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/logger"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/orders"
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

func (s *storage) AddOne(ctx context.Context, order *orders.Order) (*orders.Order, error) {
	var factUserID int64
	query := "SELECT user_id FROM ygm_order WHERE num=$1;"
	row := s.db.Pool.QueryRow(ctx, query, order.Num)
	err := row.Scan(&factUserID)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("failed to get user from DB: %w", err)
		}
	} else {
		if factUserID == order.UserID {
			return nil, db.ErrRowExist
		}
		return nil, db.ErrConflict
	}

	query = "INSERT INTO ygm_order (user_id, status_id, num, order_date)" +
		" VALUES($1, (SELECT s.id FROM ygm_order_status s WHERE s.ident=$2), $3, $4) RETURNING id;"
	row = s.db.Pool.QueryRow(ctx, query, order.UserID, order.StatusIdent, order.Num, order.Date)
	err = row.Scan(&order.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to insert order in DB: %w", err)
	}
	return order, err
}

func (s *storage) GetAll(ctx context.Context, userID int64) ([]orders.Order, error) {
	usersOrders := make([]orders.Order, 0)
	query := "SELECT ord.num, st.ident, COALESCE(op.amount, 0), ord.order_date" +
		" FROM ygm_order ord" +
		" INNER JOIN ygm_order_status st ON st.id = ord.status_id" +
		" LEFT JOIN ygm_balls_operation op ON op.order_id = ord.id" +
		" WHERE  ord.user_id = $1 ORDER BY ord.order_date;"
	rows, err := s.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders from DB: %w", err)
	}
	for rows.Next() {
		order := orders.Order{}
		if err := rows.Scan(&order.Num, &order.StatusIdent, &order.Accrual, &order.Date); err != nil {
			return nil, fmt.Errorf("failed to scan orders from rows: %w", err)
		}
		usersOrders = append(usersOrders, order)
	}
	return usersOrders, nil
}

func (s *storage) GetAllUactual(ctx context.Context) ([]orders.Order, error) {
	usersOrders := make([]orders.Order, 0)
	query := "SELECT ord.id, ord.num" +
		" FROM ygm_order ord" +
		" INNER JOIN ygm_order_status st ON st.id = ord.status_id" +
		" WHERE  st.ident NOT IN ('INVALID', 'PROCESSED') LIMIT 500;"
	rows, err := s.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders from DB: %w", err)
	}
	for rows.Next() {
		order := orders.Order{}
		if err := rows.Scan(&order.ID, &order.Ord); err != nil {
			return nil, fmt.Errorf("failed to scan orders from rows: %w", err)
		}
		usersOrders = append(usersOrders, order)
	}
	return usersOrders, nil
}

func (s *storage) Update(ctx context.Context, ordrs []orders.Order) error {
	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	queryStat := "UPDATE ygm_order" +
		" SET status_id = (SELECT s.id FROM ygm_order_status s WHERE s.ident=$1)" +
		" WHERE id = $2;"

	queryDellBalls := "DELETE FROM ygm_balls_operation" +
		" WHERE order_id = $1 AND amount > 0;"

	queryAddBalls := "INSERT INTO ygm_balls_operation (order_id, amount)" +
		" VALUES($1, $2);"

	_, err = tx.Prepare(ctx, "stmStat", queryStat)
	if err != nil {
		return fmt.Errorf("failed prepare to update order in DB: %w", err)
	}
	_, err = tx.Prepare(ctx, "stmDellBalls", queryDellBalls)
	if err != nil {
		return fmt.Errorf("failed prepare to delet old data for operation in DB: %w", err)
	}
	_, err = tx.Prepare(ctx, "stmAddBalls", queryAddBalls)
	if err != nil {
		return fmt.Errorf("failed prepare to update operation in DB: %w", err)
	}

	for _, o := range ordrs {
		result := tx.Conn().PgConn().ExecPrepared(ctx, "stmStat", [][]byte{[]byte(o.StatusIdent), []byte(strconv.FormatInt(o.ID, 10))}, nil, nil).Read()
		if result.Err != nil {
			return fmt.Errorf("failed to update order in DB: %w", err)
		}
		result = tx.Conn().PgConn().ExecPrepared(ctx, "stmDellBalls", [][]byte{[]byte(strconv.FormatInt(o.ID, 10))}, nil, nil).Read()
		if result.Err != nil {
			return fmt.Errorf("failed to delet old data for operation in DB: %w", err)
		}
		result = tx.Conn().PgConn().ExecPrepared(ctx, "stmAddBalls", [][]byte{[]byte(strconv.FormatInt(o.ID, 10)), []byte(fmt.Sprintf("%f", o.Accrual))}, nil, nil).Read()
		if result.Err != nil {
			return fmt.Errorf("failed to update operation in DB: %w", err)
		}
	}
	return tx.Commit(ctx)
}
