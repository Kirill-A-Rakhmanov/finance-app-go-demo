package dto

import (
	"finance-app-go-demo/internal/domain/enum"
	"time"

	"github.com/google/uuid"
)

type Transaction struct {
	ID                    uuid.UUID            `json:"id" validate:"required"`
	AccountID             uuid.UUID            `json:"accountId" validate:"required"`
	TransactionCategoryID *uuid.UUID           `json:"transactionCategoryId"`
	TransferID            *uuid.UUID           `json:"transferId"`
	Type                  enum.TransactionType `json:"type" validate:"required"`
	Amount                float64              `json:"amount" validate:"required"`
	Name                  *string              `json:"name"`
	CounterpartyID        *uuid.UUID           `json:"counterpartyId"`
	Date                  time.Time            `json:"date" validate:"required"`
	IsConfirmed           bool                 `json:"isConfirmed" validate:"required"`
}
