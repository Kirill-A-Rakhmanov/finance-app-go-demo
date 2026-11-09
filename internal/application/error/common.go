package error

var (
	ErrUserNotFound                = NotFound("user not found")
	ErrAccountNotFound             = NotFound("account not found")
	ErrTransactionCategoryNotFound = NotFound("transaction category not found")
	ErrCounterpartyNotFound        = NotFound("counterparty not found")
	ErrNegativeBalanceNotAllowed   = Forbidden("negative balance not allowed")
)
