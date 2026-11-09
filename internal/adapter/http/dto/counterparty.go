package dto

import "github.com/google/uuid"

type Counterparty struct {
	ID     uuid.UUID `json:"id" validate:"required"`
	UserID uuid.UUID `json:"userId" validate:"required"`
	Name   string    `json:"name" validate:"required"`
}
