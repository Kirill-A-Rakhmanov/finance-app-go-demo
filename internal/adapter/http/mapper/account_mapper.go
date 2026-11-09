package mapper

import (
	"finance-app-go-demo/internal/adapter/http/dto"
	"finance-app-go-demo/internal/domain/model"
	"finance-app-go-demo/internal/domain/valueObject"
)

func ToAccountDto(b *model.Account) *dto.Account {
	if b == nil {
		return nil
	}

	return &dto.Account{
		ID:                       b.ID,
		UserID:                   b.UserID,
		Name:                     b.Name,
		Icon:                     b.Icon,
		Currency:                 ToCurrencyDto(b.Currency),
		IsNegativeBalanceEnabled: b.IsNegativeBalanceEnabled,
		IsHidden:                 b.IsHidden,
		IsExcludedFromCount:      b.IsExcludedFromCount,
		Adjustment:               b.Adjustment.InexactFloat64(),
	}
}

func ToCurrencyDto(c valueObject.Currency) dto.Currency {
	return dto.Currency{
		Code:     string(c.Code),
		Symbol:   c.Symbol,
		Decimals: c.Decimals,
	}
}
