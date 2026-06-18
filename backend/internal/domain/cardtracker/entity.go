package cardtracker

type Card struct {
	ID               string  `json:"id"`
	Alias            string  `json:"alias"`
	Bank             string  `json:"bank"`
	Last4            string  `json:"last4"`
	PayDate          string  `json:"pay_date"`
	Limit            float64 `json:"limit"`
	EstimatedPayment float64 `json:"estimated_payment"`
	Balance          float64 `json:"balance"`
}

type Summary struct {
	TotalDebt    float64 `json:"total_debt"`
	TotalPayment float64 `json:"total_payment"`
	NextPayDate  *string `json:"next_pay_date"`
	Cards        []Card  `json:"cards"`
}
