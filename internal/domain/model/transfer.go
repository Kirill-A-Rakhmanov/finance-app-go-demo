package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Transfer struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time

	FromAccountID uuid.UUID
	ToAccountID   uuid.UUID
	FromAmount    decimal.Decimal
	ToAmount      decimal.Decimal
	Date          time.Time
}

func NewTransfer(
	fromAccountID uuid.UUID,
	toAccountID uuid.UUID,
	fromAmount decimal.Decimal,
	toAmount decimal.Decimal,
	date time.Time,
) *Transfer {
	return &Transfer{
		FromAccountID: fromAccountID,
		ToAccountID:   toAccountID,
		FromAmount:    fromAmount,
		ToAmount:      toAmount,
		Date:          date,
	}
}
