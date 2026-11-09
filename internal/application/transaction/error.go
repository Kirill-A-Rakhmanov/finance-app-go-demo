package transaction

import apperror "finance-app-go-demo/internal/application/error"

var (
	ErrTransactionNotFound           = apperror.NotFound("transaction not found")
	ErrZeroAmountNotAllowed          = apperror.Invalid("zero amount not allowed")
	ErrNegativeAmountNotAllowed      = apperror.Invalid("negative amount not allowed")
	ErrInvalidTransactionType        = apperror.Invalid("invalid transaction type")
	ErrTransferUpdateNotAllowed      = apperror.Forbidden("transfer update not allowed")
	ErrDeleteTransferNotAllowed      = apperror.Forbidden("delete transfer not allowed")
	ErrMoveToOtherCurrencyNotAllowed = apperror.Invalid("transaction can only be moved to an account with the same currency")
)
