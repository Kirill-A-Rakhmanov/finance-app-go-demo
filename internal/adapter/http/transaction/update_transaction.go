package transaction

import (
	"finance-app-go-demo/internal/adapter/http/dto"
	"finance-app-go-demo/internal/adapter/http/mapper"
	"finance-app-go-demo/internal/adapter/http/request"
	"finance-app-go-demo/internal/adapter/http/response"
	"finance-app-go-demo/internal/application/transaction"
	"finance-app-go-demo/internal/domain/enum"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type UpdateTransactionRequest struct {
	AccountID               uuid.UUID            `json:"accountId" validate:"required"`
	TransactionCategoryName *string              `json:"transactionCategoryName"`
	Amount                  decimal.Decimal      `json:"amount" validate:"required"`
	Name                    *string              `json:"name"`
	CounterpartyName        *string              `json:"counterpartyName"`
	Date                    time.Time            `json:"date" validate:"required"`
	Type                    enum.TransactionType `json:"type" validate:"required"`
	IsConfirmed             bool                 `json:"isConfirmed" validate:"required"`
}

type UpdateTransactionResponse struct {
	Transaction *dto.Transaction `json:"transaction" validate:"required"`
}

// @Summary Update transaction
// @Description Update the transaction. If accountId differs from the transaction's current account the transaction is moved there (same currency only).
// @Tags transaction
// @Accept json
// @Produce json
// @Param transactionID path string true "Transaction ID" Format(uuid)
// @Param request body UpdateTransactionRequest true "Update transaction request"
// @Success 200 {object} UpdateTransactionResponse
// @Router /transaction/{transactionID} [patch]
func (h *Handler) UpdateTransaction(c *gin.Context) {
	ctx := c.Request.Context()

	transactionID, err := uuid.Parse(c.Param("transactionID"))
	if err != nil {
		response.BadRequest(c.Writer, fmt.Errorf("invalid transactionID: %w", err))
		return
	}

	session, ok := request.CurrentSession(ctx)
	if !ok {
		response.Unauthorized(c.Writer, fmt.Errorf("user not authorized"))
		return
	}

	var req UpdateTransactionRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c.Writer, fmt.Errorf("invalid request body: %v", err))
		return
	}

	result, err := h.transactionService.UpdateTransaction(ctx, transaction.UpdateTransactionInput{
		TransactionID:           transactionID,
		UserID:                  session.UserID,
		AccountID:               req.AccountID,
		TransactionCategoryName: req.TransactionCategoryName,
		Amount:                  req.Amount,
		Name:                    req.Name,
		CounterpartyName:        req.CounterpartyName,
		Date:                    req.Date,
		Type:                    req.Type,
		IsConfirmed:             req.IsConfirmed,
	})
	if err != nil {
		response.FromError(c.Writer, err)
		return
	}

	res := UpdateTransactionResponse{
		Transaction: mapper.ToTransactionDto(result.Transaction),
	}

	response.OK(c.Writer, res)
}
