package counterparty

import (
	"finance-app-go-demo/internal/adapter/http/dto"
	"finance-app-go-demo/internal/adapter/http/mapper"
	"finance-app-go-demo/internal/adapter/http/request"
	"finance-app-go-demo/internal/adapter/http/response"
	"finance-app-go-demo/internal/application/counterparty"
	"fmt"

	"github.com/gin-gonic/gin"
)

type GetCounterpartiesResponse struct {
	Counterparties []dto.Counterparty `json:"counterparties" validate:"required"`
}

// @Summary Get counterparties
// @Description Get current user counterparties
// @Tags counterparty
// @Accept json
// @Produce json
// @Success 200 {object} GetCounterpartiesResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /counterparty/list [get]
func (h *Handler) GetCounterparties(c *gin.Context) {
	ctx := c.Request.Context()

	session, ok := request.CurrentSession(ctx)
	if !ok {
		response.Unauthorized(c.Writer, fmt.Errorf("user not authorized"))
		return
	}

	result, err := h.counterpartyService.GetCounterpartiesByUserID(ctx, counterparty.GetCounterpartiesByUserIDInput{
		UserID: session.UserID,
	})
	if err != nil {
		response.FromError(c.Writer, err)
		return
	}

	counterparties := make([]dto.Counterparty, len(result.Counterparties))
	for i, cp := range result.Counterparties {
		counterparties[i] = *mapper.ToCounterpartyDto(&cp)
	}

	res := GetCounterpartiesResponse{
		Counterparties: counterparties,
	}

	response.OK(c.Writer, res)
}
