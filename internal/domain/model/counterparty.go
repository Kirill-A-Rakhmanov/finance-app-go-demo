package model

import (
	"time"

	"github.com/google/uuid"
)

type Counterparty struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time

	UserID uuid.UUID
	Name   string
}

func NewCounterparty(
	userID uuid.UUID,
	name string,
) *Counterparty {
	return &Counterparty{
		UserID: userID,
		Name:   name,
	}
}
