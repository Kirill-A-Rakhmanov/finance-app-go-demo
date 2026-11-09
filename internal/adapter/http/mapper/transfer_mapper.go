package mapper

import (
	"finance-app-go-demo/internal/adapter/http/dto"
	"finance-app-go-demo/internal/domain/model"
)

func ToTransferDto(t *model.Transfer) *dto.Transfer {
	if t == nil {
		return nil
	}

	return &dto.Transfer{
		ID:            t.ID,
		FromAccountID: t.FromAccountID,
		ToAccountID:   t.ToAccountID,
		FromAmount:    t.FromAmount.InexactFloat64(),
		ToAmount:      t.ToAmount.InexactFloat64(),
		Date:          t.Date,
	}
}
