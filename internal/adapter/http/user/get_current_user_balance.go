package user

import (
	"finance-app-go-demo/internal/adapter/http/dto"
	"finance-app-go-demo/internal/adapter/http/mapper"
	"finance-app-go-demo/internal/adapter/http/request"
	"finance-app-go-demo/internal/adapter/http/response"
	"finance-app-go-demo/internal/application/user"
	"fmt"

	"github.com/gin-gonic/gin"
)

type CurrencyBalance struct {
	Currency dto.Currency `json:"currency" validate:"required"`
	Amount   float64      `json:"amount" validate:"required"`
}

type GetCurrentUserBalanceResponse struct {
	Balances []CurrencyBalance `json:"balances" validate:"required"`
}

// @Summary Get user balance
// @Description Get current user balances grouped by currency
// @Tags user
// @Accept json
// @Produce json
// @Success 200 {object} GetCurrentUserBalanceResponse
// @Router /user/current/balance [get]
func (h *Handler) GetCurrentUserBalance(c *gin.Context) {
	ctx := c.Request.Context()

	session, ok := request.CurrentSession(ctx)
	if !ok {
		response.Unauthorized(c.Writer, fmt.Errorf("user not authorized"))
		return
	}

	result, err := h.userService.GetBalanceByUserID(ctx, user.GetBalanceByUserIDInput{
		UserID: session.UserID,
	})
	if err != nil {
		response.FromError(c.Writer, err)
		return
	}

	balances := make([]CurrencyBalance, len(result.Balances))
	for i, balance := range result.Balances {
		balances[i] = CurrencyBalance{
			Currency: mapper.ToCurrencyDto(balance.Currency),
			Amount:   balance.Amount.InexactFloat64(),
		}
	}

	response.OK(c.Writer, GetCurrentUserBalanceResponse{
		Balances: balances,
	})
}
