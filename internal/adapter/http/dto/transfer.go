package dto

import (
	"time"

	"github.com/google/uuid"
)

type Transfer struct {
	ID            uuid.UUID `json:"id" validate:"required"`
	FromAccountID uuid.UUID `json:"fromAccountId" validate:"required"`
	ToAccountID   uuid.UUID `json:"toAccountId" validate:"required"`
	FromAmount    float64   `json:"fromAmount" validate:"required"`
	ToAmount      float64   `json:"toAmount" validate:"required"`
	Date          time.Time `json:"date" validate:"required"`
}
