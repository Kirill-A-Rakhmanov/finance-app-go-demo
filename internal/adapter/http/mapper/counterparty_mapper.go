package mapper

import (
	"finance-app-go-demo/internal/adapter/http/dto"
	"finance-app-go-demo/internal/domain/model"
)

func ToCounterpartyDto(c *model.Counterparty) *dto.Counterparty {
	if c == nil {
		return nil
	}

	return &dto.Counterparty{
		ID:     c.ID,
		UserID: c.UserID,
		Name:   c.Name,
	}
}
