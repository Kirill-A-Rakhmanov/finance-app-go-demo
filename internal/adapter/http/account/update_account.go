package account

import (
	"finance-app-go-demo/internal/adapter/http/dto"
	"finance-app-go-demo/internal/adapter/http/mapper"
	"finance-app-go-demo/internal/adapter/http/request"
	"finance-app-go-demo/internal/adapter/http/response"
	"finance-app-go-demo/internal/application/account"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UpdateAccountRequest struct {
	Name                     string `json:"name" example:"My Account" validate:"required"`
	Icon                     string `json:"icon" example:"icon" validate:"required"`
	IsNegativeBalanceEnabled bool   `json:"isNegativeBalanceEnabled" example:"false" validate:"required"`
	IsHidden                 bool   `json:"isHidden" example:"false" validate:"required"`
	IsExcludedFromCount      bool   `json:"isExcludedFromCount" example:"false" validate:"required"`
}

type UpdateAccountResponse struct {
	Account dto.Account `json:"account" validate:"required"`
}

// @Summary Update account
// @Description Update the account for current user
// @Tags account
// @Accept json
// @Produce json
// @Param accountID path string true "Account ID" Format(uuid)
// @Param request body UpdateAccountRequest true "Update account request"
// @Success 200 {object} UpdateAccountResponse
// @Router /account/{accountID} [patch]
func (h *Handler) UpdateAccount(c *gin.Context) {
	ctx := c.Request.Context()

	accountID, err := uuid.Parse(c.Param("accountID"))
	if err != nil {
		response.BadRequest(c.Writer, fmt.Errorf("invalid accountID: %w", err))
		return
	}

	session, ok := request.CurrentSession(ctx)
	if !ok {
		response.Unauthorized(c.Writer, fmt.Errorf("user not authorized"))
		return
	}

	var req UpdateAccountRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c.Writer, fmt.Errorf("invalid request body: %v", err))
		return
	}

	result, err := h.accountService.UpdateAccount(ctx, account.UpdateAccountInput{
		AccountID:                accountID,
		UserID:                   session.UserID,
		Name:                     req.Name,
		Icon:                     req.Icon,
		IsNegativeBalanceEnabled: req.IsNegativeBalanceEnabled,
		IsHidden:                 req.IsHidden,
		IsExcludedFromCount:      req.IsExcludedFromCount,
	})
	if err != nil {
		response.FromError(c.Writer, err)
		return
	}

	res := UpdateAccountResponse{
		Account: *mapper.ToAccountDto(result.Account),
	}

	response.OK(c.Writer, res)
}
