package transfer

import (
	"finance-app-go-demo/internal/adapter/http/dto"
	"finance-app-go-demo/internal/adapter/http/mapper"
	"finance-app-go-demo/internal/adapter/http/request"
	"finance-app-go-demo/internal/adapter/http/response"
	"finance-app-go-demo/internal/application/transfer"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreateTransferRequest struct {
	FromAccountID uuid.UUID       `json:"fromAccountId" required:"true" type:"uuid" validate:"required"`
	ToAccountID   uuid.UUID       `json:"toAccountId" required:"true" type:"uuid" validate:"required"`
	FromAmount    decimal.Decimal `json:"fromAmount" required:"true" type:"decimal" validate:"required"`
	ToAmount      decimal.Decimal `json:"toAmount" required:"true" type:"decimal" validate:"required"`
	Date          time.Time       `json:"date" validate:"required"`
}

type CreateTransferResponse struct {
	Transfer     dto.Transfer      `json:"transfer" validate:"required"`
	Transactions []dto.Transaction `json:"transactions" validate:"required"`
}

// @Summary Create a new transfer
// @Description Create a new transfer
// @Tags transfer
// @Accept json
// @Produce json
// @Param request body CreateTransferRequest true "Create transfer request"
// @Success 200 {object} CreateTransferResponse
// @Router /transfer [post]
func (h *Handler) CreateTransfer(c *gin.Context) {
	ctx := c.Request.Context()

	session, ok := request.CurrentSession(ctx)
	if !ok {
		response.Unauthorized(c.Writer, fmt.Errorf("user not authorized"))
		return
	}

	var req CreateTransferRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c.Writer, fmt.Errorf("invalid request body: %v", err))
		return
	}

	result, err := h.transferService.CreateTransfer(ctx, transfer.CreateTransferInput{
		UserID:        session.UserID,
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		FromAmount:    req.FromAmount,
		ToAmount:      req.ToAmount,
		Date:          req.Date,
	})
	if err != nil {
		response.FromError(c.Writer, err)
		return
	}

	transactions := make([]dto.Transaction, len(result.Transactions))
	for i, t := range result.Transactions {
		transactions[i] = *mapper.ToTransactionDto(&t)
	}

	res := CreateTransferResponse{
		Transfer:     *mapper.ToTransferDto(result.Transfer),
		Transactions: transactions,
	}

	response.OK(c.Writer, res)
}
