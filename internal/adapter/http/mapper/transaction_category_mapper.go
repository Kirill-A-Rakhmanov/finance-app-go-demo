package mapper

import (
	"finance-app-go-demo/internal/adapter/http/dto"
	"finance-app-go-demo/internal/domain/model"
)

func ToTransactionCategoryDto(tc *model.TransactionCategory) *dto.TransactionCategory {
	if tc == nil {
		return nil
	}

	return &dto.TransactionCategory{
		ID:       tc.ID,
		UserID:   tc.UserID,
		ParentID: tc.ParentID,
		Name:     tc.Name,
		Icon:     tc.Icon,
		Color:    tc.Color,
		Type:     tc.Type,
	}
}
