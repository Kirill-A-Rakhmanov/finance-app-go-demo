package transaction_category

import (
	"finance-app-go-demo/internal/adapter/http/dto"
	"finance-app-go-demo/internal/adapter/http/mapper"
	"finance-app-go-demo/internal/adapter/http/request"
	"finance-app-go-demo/internal/adapter/http/response"
	"finance-app-go-demo/internal/application/transaction_category"
	"finance-app-go-demo/internal/domain/enum"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CreateTransactionCategoryRequest struct {
	ParentID *uuid.UUID `json:"parentId" type:"uuid"`
	Name     string     `json:"name" example:"Custom category" validate:"required"`
	Icon     string     `json:"icon" example:"icon" validate:"required"`
	Color    string     `json:"color" example:"#ff0000" validate:"required"`
	Type     string     `json:"type" example:"expense" validate:"required"`
}

type CreateTransactionCategoryResponse struct {
	TransactionCategory dto.TransactionCategory `json:"transactionCategory" validate:"required"`
}

// @Summary Create transaction category
// @Description Create new transaction category for current user
// @Tags transaction category
// @Accept json
// @Produce json
// @Param request body CreateTransactionCategoryRequest true "Create transaction category request"
// @Success 200 {object} CreateTransactionCategoryResponse
// @Router /transaction-category [post]
func (h *Handler) CreateTransactionCategory(c *gin.Context) {
	ctx := c.Request.Context()

	session, ok := request.CurrentSession(ctx)
	if !ok {
		response.Unauthorized(c.Writer, fmt.Errorf("user not authorized"))
		return
	}

	var req CreateTransactionCategoryRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c.Writer, fmt.Errorf("invalid request body: %v", err))
		return
	}

	result, err := h.transactionCategoryService.CreateTransactionCategory(ctx, transaction_category.CreateTransactionCategoryInput{
		UserID:   session.UserID,
		ParentID: req.ParentID,
		Name:     req.Name,
		Icon:     req.Icon,
		Color:    req.Color,
		Type:     enum.TransactionCategoryType(req.Type),
	})
	if err != nil {
		response.FromError(c.Writer, err)
		return
	}

	res := CreateTransactionCategoryResponse{
		TransactionCategory: *mapper.ToTransactionCategoryDto(result.TransactionCategory),
	}

	response.OK(c.Writer, res)
}
