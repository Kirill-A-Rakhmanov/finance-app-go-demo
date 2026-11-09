package dto

import (
	"github.com/google/uuid"
)

type User struct {
	ID         uuid.UUID `json:"id" validate:"required"`
	Email      string    `json:"email" validate:"required"`
	FirstName  string    `json:"firstName" validate:"required"`
	LastName   string    `json:"lastName" validate:"required"`
	MiddleName *string   `json:"middleName"`
}
