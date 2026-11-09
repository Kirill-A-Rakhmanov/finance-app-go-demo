package account

import (
	"finance-app-go-demo/internal/adapter/http/dto"
	"finance-app-go-demo/internal/adapter/http/mapper"
	"finance-app-go-demo/internal/adapter/http/request"
	"finance-app-go-demo/internal/adapter/http/response"
	"finance-app-go-demo/internal/application/account"
	"fmt"

	"github.com/gin-gonic/gin"
)

type GetAccountsByUserIDResponse struct {
	Accounts []dto.Account `json:"accounts" validate:"required"`
}

// @Summary Get accounts
// @Description Get current user accounts
// @Tags account
// @Accept json
// @Produce json
// @Success 200 {object} GetAccountsByUserIDResponse
// @Router /account/list [get]
func (h *Handler) GetAccountsByUserID(c *gin.Context) {
	ctx := c.Request.Context()

	session, ok := request.CurrentSession(ctx)
	if !ok {
		response.Unauthorized(c.Writer, fmt.Errorf("user not authorized"))
		return
	}

	result, err := h.accountService.GetAccountsByUserId(ctx, account.GetAccountsByUserIdInput{
		UserID: session.UserID,
	})
	if err != nil {
		response.FromError(c.Writer, err)
		return
	}

	accounts := make([]dto.Account, len(result.Accounts))
	for i, account := range result.Accounts {
		accounts[i] = *mapper.ToAccountDto(&account)
	}

	res := GetAccountsByUserIDResponse{
		Accounts: accounts,
	}

	response.OK(c.Writer, res)
}
