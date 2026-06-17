package finance

type RecentTransaction struct {
	ID          string  `json:"id"`
	Amount      float64 `json:"amount"`
	Type        string  `json:"type"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
	Date        string  `json:"date"`
}

type Summary struct {
	Balance           float64             `json:"balance"`
	MonthIncome       float64             `json:"month_income"`
	MonthExpenses     float64             `json:"month_expenses"`
	RecentTransactions []RecentTransaction `json:"recent_transactions"`
}
