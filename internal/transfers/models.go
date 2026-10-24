package transfers

type AccountInfo struct {
	Balance  string
	Currency string
}

type TransferInfo struct {
	FromAccount string
	ToAccount   string
	Amount      string
	IKey        string
}

type Transfer struct {
	ID          string
	FromAccount string
	ToAccount   string
	Amount      string
}
