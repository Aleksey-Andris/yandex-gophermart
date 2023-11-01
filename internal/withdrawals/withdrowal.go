package withdrawals

import "time"

type Withdrowal struct {
	Ord    string    `json:"order"`
	Amount float64   `json:"sum"`
	Data   time.Time `json:"processed_at"`
}
