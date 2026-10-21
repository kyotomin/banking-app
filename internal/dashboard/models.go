package dashboard

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
