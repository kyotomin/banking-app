package dashboard

import "time"

type DashboardStats struct {
	Owner      string `json:"owner"`
	Balance    string `json:"balance"`
	Currency   string `json:"currency"`
	TotalSpent string `json:"total_spent"`
}

type Account struct {
	Owner    string
	Balance  string
	Currency string
}

type TotalSpent struct {
	Amount string
}

type Entry struct {
	ID         string
	TransferID string
	Amount     string
	Type       string
	Category   string
	CreatedAt  time.Time
}

type MonthlyStats struct {
	TotalIncome string  `json:"total_income"`
	TotalSpent  string  `json:"total_spent"`
	Entries     []Entry `json:"entries"`
}
