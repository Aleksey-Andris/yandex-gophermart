package storages

import (
	"context"
	"errors"

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
	var factUserId int64
	query := "SELECT user_id FROM ygm_order WHERE num=$1;"
	row := s.db.Pool.QueryRow(ctx, query, order.Num)
	err := row.Scan(&factUserId)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
	} else {
		if factUserId == order.UserID {
			return nil, db.ErrRowExist
		}
		return nil, db.ErrConflict
	}

	query = "INSERT INTO ygm_order (user_id, status_id, num, order_date)" +
		" VALUES($1, (SELECT s.id FROM ygm_order_status s WHERE s.ident=$2), $3, $4) RETURNING id;"
	row = s.db.Pool.QueryRow(ctx, query, order.UserID, order.StatusIdent, order.Num, order.Date)
	err = row.Scan(&order.ID)
	if err != nil {
		return nil, err
	}
	return order, err
}

func (s *storage) GetAll(ctx context.Context, userID int64) ([]orders.Order, error) {
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

	usersOrders := make([]orders.Order, 0)
	query = "SELECT ord.num, st.ident, op.amount, ord.order_date" +
		" FROM ygm_order ord" +
		" INNER JOIN ygm_order_status st ON st.id = ord.status_id" +
		" LEFT JOIN ygm_balls_operation op ON op.order_id = ord.id" +
		" WHERE  ord.user_id = $1 ORDER BY ord.order_date;"
	rows, err := s.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		order := orders.Order{}
		if err := rows.Scan(&order.Ord, &order.StatusIdent, &order.Accrual, &order.Date); err != nil {
			return nil, err
		}
		usersOrders = append(usersOrders, order)
	}
	return usersOrders, nil
}
