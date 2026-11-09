package transaction_category

import (
	"finance-app-go-demo/internal/adapter/http/dto"
	"finance-app-go-demo/internal/adapter/http/mapper"
	"finance-app-go-demo/internal/adapter/http/request"
	"finance-app-go-demo/internal/adapter/http/response"
	"finance-app-go-demo/internal/application/transaction_category"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UpdateTransactionCategoryRequest struct {
	ParentID *uuid.UUID `json:"parentId" type:"uuid"`
	Name     string     `json:"name" example:"Custom category" validate:"required"`
	Icon     string     `json:"icon" example:"icon" validate:"required"`
	Color    string     `json:"color" example:"#ff0000" validate:"required"`
}

type UpdateTransactionCategoryResponse struct {
	TransactionCategory dto.TransactionCategory `json:"transactionCategory" validate:"required"`
}

// @Summary Update transaction category
// @Description Update the transaction category for current user
// @Tags transaction category
// @Accept json
// @Produce json
// @Param transactionCategoryID path string true "Transaction category ID" Format(uuid)
// @Param request body UpdateTransactionCategoryRequest true "Update transaction category request"
// @Success 200 {object} UpdateTransactionCategoryResponse
// @Router /transaction-category/{transactionCategoryID} [patch]
func (h *Handler) UpdateTransactionCategory(c *gin.Context) {
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

	var req UpdateTransactionCategoryRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c.Writer, fmt.Errorf("invalid request body: %v", err))
		return
	}

	result, err := h.transactionCategoryService.UpdateTransactionCategory(ctx, transaction_category.UpdateTransactionCategoryInput{
		UserID:                session.UserID,
		TransactionCategoryID: transactionCategoryID,
		ParentID:              req.ParentID,
		Name:                  req.Name,
		Icon:                  req.Icon,
		Color:                 req.Color,
	})
	if err != nil {
		response.FromError(c.Writer, err)
		return
	}

	res := UpdateTransactionCategoryResponse{
		TransactionCategory: *mapper.ToTransactionCategoryDto(result.TransactionCategory),
	}

	response.OK(c.Writer, res)
}
