package user

import apperror "finance-app-go-demo/internal/application/error"

var ErrInvalidCurrentPassword = apperror.Forbidden("invalid current password")
