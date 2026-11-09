package mapper

import (
	"finance-app-go-demo/internal/adapter/http/dto"
	"finance-app-go-demo/internal/domain/model"
)

func ToTransactionDto(t *model.Transaction) *dto.Transaction {
	if t == nil {
		return nil
	}

	return &dto.Transaction{
		ID:                    t.ID,
		AccountID:             t.AccountID,
		TransactionCategoryID: t.TransactionCategoryID,
		TransferID:            t.TransferID,
		Type:                  t.Type,
		Amount:                t.Amount.InexactFloat64(),
		Name:                  t.Name,
		CounterpartyID:        t.CounterpartyID,
		Date:                  t.Date,
		IsConfirmed:           t.IsConfirmed,
	}
}
