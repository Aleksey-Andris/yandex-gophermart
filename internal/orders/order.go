package orders

import "time"

const (
	StatusNew        = "NEW"
	StatusProcessing = "PROCESSING"
	StatusInvalid    = "INVALID"
	StatusProcessed  = "PROCESSED"
)

type Order struct {
	ID          int64     `json:"-"`
	Num         int64     `json:"-"`
	Ord         string    `json:"order"`
	StatusIdent string    `json:"status"`
	UserID      int64     `json:"-"`
	Date        time.Time `json:"uploaded_at"`
	Accrual     *float64  `json:"accrual,omitempty"`
}
