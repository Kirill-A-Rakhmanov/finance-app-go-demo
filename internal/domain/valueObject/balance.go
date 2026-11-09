package valueObject

import "github.com/shopspring/decimal"

type Balance struct {
	Currency Currency
	Amount   decimal.Decimal
}
