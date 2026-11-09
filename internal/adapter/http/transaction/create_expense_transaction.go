package transaction

import (
	"finance-app-go-demo/internal/adapter/http/dto"
	"finance-app-go-demo/internal/adapter/http/mapper"
	"finance-app-go-demo/internal/adapter/http/request"
	"finance-app-go-demo/internal/adapter/http/response"
	"finance-app-go-demo/internal/application/transaction"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreateExpenseTransactionRequest struct {
	AccountID               uuid.UUID       `json:"accountId" type:"uuid" validate:"required"`
	TransactionCategoryName *string         `json:"transactionCategoryName"`
	CounterpartyName        *string         `json:"counterpartyName"`
	Amount                  decimal.Decimal `json:"amount" example:"0.00" validate:"required"`
	Name                    *string         `json:"name" example:"New transaction"`
	Date                    time.Time       `json:"date" validate:"required"`
}

type CreateExpenseTransactionResponse struct {
	Transaction dto.Transaction `json:"transaction" validate:"required"`
}

// @Summary Create a new transaction
// @Description Create a new transaction
// @Tags transaction
// @Accept json
// @Produce json
// @Param request body CreateExpenseTransactionRequest true "Create expense transaction request"
// @Success 200 {object} CreateExpenseTransactionResponse
// @Router /transaction/expense [post]
func (h *Handler) CreateExpenseTransaction(c *gin.Context) {
	ctx := c.Request.Context()

	session, ok := request.CurrentSession(ctx)
	if !ok {
		response.Unauthorized(c.Writer, fmt.Errorf("user not authorized"))
		return
	}

	var req CreateExpenseTransactionRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c.Writer, fmt.Errorf("invalid request body: %v", err))
		return
	}

	result, err := h.transactionService.CreateExpenseTransaction(ctx, transaction.CreateExpenseTransactionInput{
		UserID:                  session.UserID,
		AccountID:               req.AccountID,
		TransactionCategoryName: req.TransactionCategoryName,
		Amount:                  req.Amount,
		Name:                    req.Name,
		CounterpartyName:        req.CounterpartyName,
		Date:                    req.Date,
	})
	if err != nil {
		response.FromError(c.Writer, err)
		return
	}

	res := CreateExpenseTransactionResponse{
		Transaction: *mapper.ToTransactionDto(result.Transaction),
	}

	response.OK(c.Writer, res)
}
