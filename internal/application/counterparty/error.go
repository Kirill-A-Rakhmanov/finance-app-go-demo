package counterparty

import apperror "finance-app-go-demo/internal/application/error"

var (
	ErrCounterpartyAlreadyExists = apperror.Conflict("counterparty with this name already exists")
)
