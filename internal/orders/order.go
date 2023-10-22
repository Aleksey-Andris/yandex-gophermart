package orders

import "time"

const (
	StatusNew        = "NEW"
	StatusProcessing = "PROCESSING"
	StatusInvalid    = "INVALID"
	StatusProcessed  = "PROCESSED"
)

type Order struct {
	ID          int64     `db:"id"`
	UserID      int64     `db:"user_id"`
	StatusIdent string    `db:"status"`
	Num         int64     `db:"num"`
	Date        time.Time `db:"order_date"`
}
