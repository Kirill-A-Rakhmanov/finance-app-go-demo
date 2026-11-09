package model

import (
	"finance-app-go-demo/internal/domain/enum"
	"time"

	"github.com/google/uuid"
)

const (
	DefaultTransactionCategoryIcon  = "tag"
	DefaultTransactionCategoryColor = "#3B82F6"
)

type TransactionCategory struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time

	UserID    uuid.UUID
	ParentID  *uuid.UUID
	Name      string
	Slug      *string
	Icon      string
	Color     string
	IsDefault bool
	Type      enum.TransactionCategoryType
}

func NewTransactionCategory(
	userID uuid.UUID,
	parentID *uuid.UUID,
	name string,
	icon string,
	color string,
	type_ enum.TransactionCategoryType,
) *TransactionCategory {
	return &TransactionCategory{
		UserID:    userID,
		ParentID:  parentID,
		Name:      name,
		Slug:      nil,
		Icon:      icon,
		Color:     color,
		IsDefault: false,
		Type:      type_,
	}
}
