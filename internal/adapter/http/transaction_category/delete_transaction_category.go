package transaction_category

import (
	"finance-app-go-demo/internal/adapter/http/request"
	"finance-app-go-demo/internal/adapter/http/response"
	"finance-app-go-demo/internal/application/transaction_category"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type DeleteTransactionCategoryResponse struct {
	Message string `json:"message" validate:"required"`
}

// @Summary Delete transaction category
// @Description Delete transaction category by id
// @Tags transaction category
// @Accept json
// @Produce json
// @Param transactionCategoryID path string true "Transaction category ID" Format(uuid)
// @Success 200 {object} DeleteTransactionCategoryResponse
// @Router /transaction-category/{transactionCategoryID} [delete]
func (h *Handler) DeleteTransactionCategory(c *gin.Context) {
	ctx := c.Request.Context()

	transactionCategoryID, err := uuid.Parse(c.Param("transactionCategoryID"))
	if err != nil {
		response.BadRequest(c.Writer, fmt.Errorf("invalid transactionCategoryID: %w", err))
		return
	}

	session, ok := request.CurrentSession(ctx)
	if !ok {
		response.Unauthorized(c.Writer, fmt.Errorf("user not authorized"))
		return
	}

	result, err := h.transactionCategoryService.DeleteTransactionCategory(ctx, transaction_category.DeleteTransactionCategoryInput{
		TransactionCategoryID: transactionCategoryID,
		UserID:                session.UserID,
	})
	if err != nil {
		response.FromError(c.Writer, err)
		return
	}

	res := DeleteTransactionCategoryResponse{
		Message: result.Message,
	}

	response.OK(c.Writer, res)
}
