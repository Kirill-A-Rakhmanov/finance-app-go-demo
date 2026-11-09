package transaction_category

import (
	"finance-app-go-demo/internal/adapter/http/dto"
	"finance-app-go-demo/internal/adapter/http/mapper"
	"finance-app-go-demo/internal/adapter/http/request"
	"finance-app-go-demo/internal/adapter/http/response"
	"finance-app-go-demo/internal/application/transaction_category"
	"fmt"

	"github.com/gin-gonic/gin"
)

type GetTransactionCategoriesResponse struct {
	TransactionCategories []dto.TransactionCategory `json:"transactionCategories" validate:"required"`
}

// @Summary Get transaction categories
// @Description Get current user transaction categories
// @Tags transaction category
// @Accept json
// @Produce json
// @Success 200 {object} GetTransactionCategoriesResponse
// @Router /transaction-category/list [get]
func (h *Handler) GetTransactionCategories(c *gin.Context) {
	ctx := c.Request.Context()

	session, ok := request.CurrentSession(ctx)
	if !ok {
		response.Unauthorized(c.Writer, fmt.Errorf("user not authorized"))
		return
	}

	result, err := h.transactionCategoryService.GetTransactionCategoriesByUserID(ctx, transaction_category.GetTransactionCategoriesByUserIDInput{
		UserID: session.UserID,
	})
	if err != nil {
		response.FromError(c.Writer, err)
		return
	}

	categories := make([]dto.TransactionCategory, len(result.TransactionCategories))
	for i, category := range result.TransactionCategories {
		categories[i] = *mapper.ToTransactionCategoryDto(&category)
	}

	res := GetTransactionCategoriesResponse{
		TransactionCategories: categories,
	}

	response.OK(c.Writer, res)
}
