package counterparty

import (
	"finance-app-go-demo/internal/adapter/http/dto"
	"finance-app-go-demo/internal/adapter/http/mapper"
	"finance-app-go-demo/internal/adapter/http/request"
	"finance-app-go-demo/internal/adapter/http/response"
	"finance-app-go-demo/internal/application/counterparty"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UpdateCounterpartyRequest struct {
	Name string `json:"name" example:"Supermarket" validate:"required"`
}

type UpdateCounterpartyResponse struct {
	Counterparty dto.Counterparty `json:"counterparty" validate:"required"`
}

// @Summary Update counterparty
// @Description Rename the counterparty for current user
// @Tags counterparty
// @Accept json
// @Produce json
// @Param counterpartyID path string true "Counterparty ID" Format(uuid)
// @Param request body UpdateCounterpartyRequest true "Update counterparty request"
// @Success 200 {object} UpdateCounterpartyResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 422 {object} response.ValidationErrorResponse
// @Router /counterparty/{counterpartyID} [patch]
func (h *Handler) UpdateCounterparty(c *gin.Context) {
	ctx := c.Request.Context()

	counterpartyID, err := uuid.Parse(c.Param("counterpartyID"))
	if err != nil {
		response.BadRequest(c.Writer, fmt.Errorf("invalid counterpartyID: %w", err))
		return
	}

	session, ok := request.CurrentSession(ctx)
	if !ok {
		response.Unauthorized(c.Writer, fmt.Errorf("user not authorized"))
		return
	}

	var req UpdateCounterpartyRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c.Writer, fmt.Errorf("invalid request body: %v", err))
		return
	}

	result, err := h.counterpartyService.UpdateCounterparty(ctx, counterparty.UpdateCounterpartyInput{
		UserID:         session.UserID,
		CounterpartyID: counterpartyID,
		Name:           req.Name,
	})
	if err != nil {
		response.FromError(c.Writer, err)
		return
	}

	res := UpdateCounterpartyResponse{
		Counterparty: *mapper.ToCounterpartyDto(result.Counterparty),
	}

	response.OK(c.Writer, res)
}
