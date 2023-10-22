package storages

import (
	"context"
	"errors"
	"fmt"

	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/db"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/logger"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/orders"
	"github.com/jackc/pgx/v5"
)

const (
	orderTable  = "ygm_order"
	orderNum    = "num"
	userID      = "user_id"
	statusID    = "status_id"
	statusIdent = "ident"
	orderDate   = "order_date"
	statusTable = "ygm_order_status"
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
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s=$1;", userID, orderTable, orderNum)
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

	query = fmt.Sprintf("INSERT INTO %s (%s, %s, %s, %s) VALUES($1, (SELECT s.id FROM %s s WHERE s.%s=$2), $3, $4) RETURNING id;",
		orderTable, userID, statusID, orderNum, orderDate, statusTable, statusIdent)
	row = s.db.Pool.QueryRow(ctx, query, order.UserID, order.StatusIdent, order.Num, order.Date)
	err = row.Scan(&order.ID)
	if err != nil {
		return nil, err
	}
	return order, err
}

func (s *storage) GetAll(ctx context.Context, order *orders.Order) (*orders.Order, error) {
	return nil, nil
}
