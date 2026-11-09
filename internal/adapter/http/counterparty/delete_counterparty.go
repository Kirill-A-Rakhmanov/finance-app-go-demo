package counterparty

import (
	"finance-app-go-demo/internal/adapter/http/request"
	"finance-app-go-demo/internal/adapter/http/response"
	"finance-app-go-demo/internal/application/counterparty"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type DeleteCounterpartyResponse struct {
	Message string `json:"message" validate:"required"`
}

// @Summary Delete counterparty
// @Description Delete counterparty by id; its transactions are kept without a counterparty
// @Tags counterparty
// @Accept json
// @Produce json
// @Param counterpartyID path string true "Counterparty ID" Format(uuid)
// @Success 200 {object} DeleteCounterpartyResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /counterparty/{counterpartyID} [delete]
func (h *Handler) DeleteCounterparty(c *gin.Context) {
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

	result, err := h.counterpartyService.DeleteCounterparty(ctx, counterparty.DeleteCounterpartyInput{
		CounterpartyID: counterpartyID,
		UserID:         session.UserID,
	})
	if err != nil {
		response.FromError(c.Writer, err)
		return
	}

	res := DeleteCounterpartyResponse{
		Message: result.Message,
	}

	response.OK(c.Writer, res)
}
