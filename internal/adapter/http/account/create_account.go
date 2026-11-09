package account

import (
	"finance-app-go-demo/internal/adapter/http/dto"
	"finance-app-go-demo/internal/adapter/http/mapper"
	"finance-app-go-demo/internal/adapter/http/request"
	"finance-app-go-demo/internal/adapter/http/response"
	"finance-app-go-demo/internal/application/account"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type CreateAccountRequest struct {
	Name                     string          `json:"name" example:"My Account" validate:"required"`
	Icon                     string          `json:"icon" example:"icon" validate:"required"`
	Currency                 string          `json:"currency" example:"USD" validate:"required"`
	IsNegativeBalanceEnabled bool            `json:"isNegativeBalanceEnabled" example:"false" validate:"required"`
	IsHidden                 bool            `json:"isHidden" example:"false" validate:"required"`
	IsExcludedFromCount      bool            `json:"isExcludedFromCount" example:"false" validate:"required"`
	InitialBalance           decimal.Decimal `json:"initialBalance" example:"0.00" validate:"required"`
}

type CreateAccountResponse struct {
	Account dto.Account `json:"account" validate:"required"`
}

// @Summary Create account
// @Description Create new account for current user
// @Tags account
// @Accept json
// @Produce json
// @Param request body CreateAccountRequest true "Create account request"
// @Success 200 {object} CreateAccountResponse
// @Router /account [post]
func (h *Handler) CreateAccount(c *gin.Context) {
	ctx := c.Request.Context()

	session, ok := request.CurrentSession(ctx)
	if !ok {
		response.Unauthorized(c.Writer, fmt.Errorf("user not authorized"))
		return
	}

	var req CreateAccountRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c.Writer, fmt.Errorf("invalid request body: %v", err))
		return
	}

	result, err := h.accountService.CreateAccount(ctx, account.CreateAccountInput{
		UserID:                   session.UserID,
		Name:                     req.Name,
		Icon:                     req.Icon,
		Currency:                 req.Currency,
		IsNegativeBalanceEnabled: req.IsNegativeBalanceEnabled,
		IsHidden:                 req.IsHidden,
		IsExcludedFromCount:      req.IsExcludedFromCount,
		InitialBalance:           req.InitialBalance,
	})
	if err != nil {
		response.FromError(c.Writer, err)
		return
	}

	res := CreateAccountResponse{
		Account: *mapper.ToAccountDto(result.Account),
	}

	response.OK(c.Writer, res)
}
