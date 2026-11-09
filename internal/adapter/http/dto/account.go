package dto

import (
	"github.com/google/uuid"
)

type Currency struct {
	Code     string `json:"code" example:"USD" validate:"required"`
	Symbol   string `json:"symbol" example:"$" validate:"required"`
	Decimals int    `json:"decimals" example:"2" validate:"required"`
}

type Account struct {
	ID                       uuid.UUID `json:"id" validate:"required"`
	UserID                   uuid.UUID `json:"userId" validate:"required"`
	Name                     string    `json:"name" validate:"required"`
	Icon                     string    `json:"icon" validate:"required"`
	Currency                 Currency  `json:"currency" validate:"required"`
	IsNegativeBalanceEnabled bool      `json:"isNegativeBalanceEnabled" validate:"required"`
	IsHidden                 bool      `json:"isHidden" validate:"required"`
	IsExcludedFromCount      bool      `json:"isExcludedFromCount" validate:"required"`
	Adjustment               float64   `json:"adjustment" validate:"required"`
}
