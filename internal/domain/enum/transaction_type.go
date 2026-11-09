package enum

type TransactionType string

const (
	TransactionExpense TransactionType = "expense"
	TransactionIncome  TransactionType = "income"
)
