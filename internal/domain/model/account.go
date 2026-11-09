package model

import (
	"finance-app-go-demo/internal/domain/valueObject"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Account struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time

	UserID                   uuid.UUID
	Name                     string
	Icon                     string
	Currency                 valueObject.Currency
	IsNegativeBalanceEnabled bool
	IsHidden                 bool
	IsExcludedFromCount      bool
	Adjustment               decimal.Decimal
}

func NewAccount(
	userID uuid.UUID,
	name string,
	icon string,
	currency valueObject.Currency,
	isNegativeBalanceEnabled bool,
	isHidden bool,
	isExcludedFromCount bool,
	adjustment decimal.Decimal,
) *Account {
	return &Account{
		UserID:                   userID,
		Name:                     name,
		Icon:                     icon,
		Currency:                 currency,
		IsNegativeBalanceEnabled: isNegativeBalanceEnabled,
		IsHidden:                 isHidden,
		IsExcludedFromCount:      isExcludedFromCount,
		Adjustment:               adjustment,
	}
}

func AccountBalanceLockKey(accountID uuid.UUID) string {
	return "account-balance:" + accountID.String()
}
