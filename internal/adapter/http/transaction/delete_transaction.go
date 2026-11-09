package transaction

import (
	"finance-app-go-demo/internal/adapter/http/request"
	"finance-app-go-demo/internal/adapter/http/response"
	"finance-app-go-demo/internal/application/transaction"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type DeleteTransactionResponse struct {
	Message string          `json:"message" validate:"required"`
	Balance decimal.Decimal `json:"balance" validate:"required"`
}

// @Summary Delete transaction
// @Description Delete transaction by id
// @Tags transaction
// @Accept json
// @Produce json
// @Param transactionID path string true "Transaction ID" Format(uuid)
// @Success 200 {object} DeleteTransactionResponse
// @Router /transaction/{transactionID} [delete]
func (h *Handler) DeleteTransaction(c *gin.Context) {
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

	result, err := h.transactionService.DeleteTransaction(ctx, transaction.DeleteTransactionInput{
		TransactionID: transactionID,
		UserID:        session.UserID,
	})
	if err != nil {
		response.FromError(c.Writer, err)
		return
	}

	res := DeleteTransactionResponse{
		Message: result.Message,
		Balance: result.Balance,
	}

	response.OK(c.Writer, res)
}
