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

type GetAccountByIDResponse struct {
	Account dto.Account `json:"account" validate:"required"`
}

// @Summary Get account by id
// @Description Get account by id if it belongs to the current user
// @Tags account
// @Accept json
// @Produce json
// @Param accountID path string true "Account ID" Format(uuid)
// @Success 200 {object} GetAccountByIDResponse
// @Router /account/{accountID} [get]
func (h *Handler) GetAccountByID(c *gin.Context) {
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

	result, err := h.accountService.GetAccountById(ctx, account.GetAccountByIdInput{
		UserID:    session.UserID,
		AccountID: accountID,
	})
	if err != nil {
		response.FromError(c.Writer, err)
		return
	}

	res := GetAccountByIDResponse{
		Account: *mapper.ToAccountDto(&result.Account),
	}

	response.OK(c.Writer, res)
}
