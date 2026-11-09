package transfer

import apperror "finance-app-go-demo/internal/application/error"

var (
	ErrTransferToSameAccountNotAllowed = apperror.Invalid("transfer to same account not allowed")
	ErrAmountMismatchForSameCurrency   = apperror.Invalid("from and to amounts must match for same currency transfer")
	ErrTransferNotFound                = apperror.NotFound("transfer not found")
)
