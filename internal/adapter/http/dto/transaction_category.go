package dto

import (
	"finance-app-go-demo/internal/domain/enum"

	"github.com/google/uuid"
)

type TransactionCategory struct {
	ID       uuid.UUID                    `json:"id" validate:"required"`
	UserID   uuid.UUID                    `json:"userId" validate:"required"`
	ParentID *uuid.UUID                   `json:"parentId"`
	Name     string                       `json:"name" validate:"required"`
	Icon     string                       `json:"icon" validate:"required"`
	Color    string                       `json:"color" validate:"required"`
	Type     enum.TransactionCategoryType `json:"type" validate:"required"`
}
