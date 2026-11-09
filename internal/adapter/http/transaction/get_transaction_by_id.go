package transaction

import (
	"finance-app-go-demo/internal/adapter/http/dto"
	"finance-app-go-demo/internal/adapter/http/mapper"
	"finance-app-go-demo/internal/adapter/http/request"
	"finance-app-go-demo/internal/adapter/http/response"
	"finance-app-go-demo/internal/application/transaction"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type GetTransactionByIdResponse struct {
	Transaction dto.Transaction `json:"transaction" validate:"required"`
}

// @Summary Get transaction
// @Description Get transaction by ID
// @Tags transaction
// @Accept json
// @Produce json
// @Param transactionID path string true "Transaction ID" Format(uuid)
// @Success 200 {object} GetTransactionByIdResponse
// @Router /transaction/{transactionID} [get]
func (h *Handler) GetTransactionByID(c *gin.Context) {
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

	result, err := h.transactionService.GetTransactionById(ctx, transaction.GetTrainsactionByIdInput{
		UserID:        session.UserID,
		TransactionID: transactionID,
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
