package validation

import (
	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/domain/model"
	"fmt"

	"github.com/shopspring/decimal"
)

func Amount(v *apperror.ValidationError, field string, amount decimal.Decimal) {
	if !amount.IsPositive() {
		v.Add(field, "must be greater than zero")
		return
	}

	v.AddIf(amount.GreaterThan(model.AmountMax), field, fmt.Sprintf("must not exceed %s", model.AmountMax))
	v.AddIf(!amount.Round(model.AmountMaxScale).Equal(amount), field, fmt.Sprintf("must have at most %d decimal places", model.AmountMaxScale))
}
