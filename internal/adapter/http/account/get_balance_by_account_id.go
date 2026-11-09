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

type GetBalanceByAccountIdResponse struct {
	AccountID uuid.UUID    `json:"accountID" validate:"required"`
	Balance   float64      `json:"balance" validate:"required"`
	Currency  dto.Currency `json:"currency" validate:"required"`
}

// @Summary Get balance
// @Description Get balance by account id
// @Tags account
// @Accept json
// @Produce json
// @Param accountID path string true "Account ID" Format(uuid)
// @Success 200 {object} GetBalanceByAccountIdResponse
// @Router /account/{accountID}/balance [get]
func (h *Handler) GetBalanceByAccountID(c *gin.Context) {
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

	result, err := h.accountService.GetBalanceByAccountId(ctx, account.GetBalanceByAccountIdInput{
		UserID:    session.UserID,
		AccountID: accountID,
	})
	if err != nil {
		response.FromError(c.Writer, err)
		return
	}

	res := GetBalanceByAccountIdResponse{
		AccountID: accountID,
		Balance:   result.Balance.InexactFloat64(),
		Currency:  mapper.ToCurrencyDto(result.Currency),
	}

	response.OK(c.Writer, res)
}
