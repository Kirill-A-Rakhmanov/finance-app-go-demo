package transaction_category

import apperror "finance-app-go-demo/internal/application/error"

var (
	ErrTransactionCategoryTypeMismatch  = apperror.Invalid("transaction category type mismatch")
	ErrParentCategoryAlreadyHasParent   = apperror.Invalid("parent category already has parent")
	ErrTransactionCategoryHasChildren   = apperror.Invalid("transaction category has children")
	ErrTransactionCategoryAlreadyExists = apperror.Conflict("transaction category with this name already exists")
	ErrTransactionCategorySelfParent    = apperror.Invalid("transaction category cannot be parent to itself")
)
