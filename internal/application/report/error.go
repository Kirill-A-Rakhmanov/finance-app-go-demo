package report

import apperror "finance-app-go-demo/internal/application/error"

var (
	ErrAccountCurrencyMismatch     = apperror.Invalid("account currency does not match report currency")
	ErrSubcategoryFilterNotAllowed = apperror.Invalid("subcategories cannot be selected when subcategories are grouped")
)
