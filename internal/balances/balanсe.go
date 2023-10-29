package balances

const (
	ResultNotEnough string = "Not enough balls"
	ResultOK        string = "success"
)

type Balance struct {
	Current    float64     `json:"current"`
	Withdrawn  float64     `json:"withdrawn"`
	Operations []Operation `json:"operations,omitempty"`
}

type Operation struct {
	Ord    string  `json:"order"`
	Amount float64 `json:"sum"`
	Result string  `json:"-"`
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
