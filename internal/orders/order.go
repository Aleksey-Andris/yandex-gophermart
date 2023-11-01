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
	Ord         string    `json:"order,omitempty"`
	Num         string    `json:"number,omitempty"`
	StatusIdent string    `json:"status,omitempty"`
	UserID      int64     `json:"-"`
	Date        time.Time `json:"uploaded_at,omitempty"`
	Accrual     float64   `json:"accrual,omitempty"`
}

func ValidLoon(number int) bool {
	return (number%10+checksum(number/10))%10 == 0
}

func checksum(number int) int {
	var luhn int
	for i := 0; number > 0; i++ {
		cur := number % 10
		if i%2 == 0 {
			cur = cur * 2
			if cur > 9 {
				cur = cur%10 + cur/10
			}
		}
		luhn += cur
		number = number / 10
	}
	return luhn % 10
}
