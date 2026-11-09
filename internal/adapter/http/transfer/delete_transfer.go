package transfer

import (
	"finance-app-go-demo/internal/adapter/http/request"
	"finance-app-go-demo/internal/adapter/http/response"
	"finance-app-go-demo/internal/application/transfer"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type DeleteTransferResponse struct {
	Message string `json:"message" validate:"required"`
}

// @Summary Delete transfer
// @Description Delete transfer by id
// @Tags transfer
// @Accept json
// @Produce json
// @Param transferID path string true "Transfer ID" Format(uuid)
// @Success 200 {object} DeleteTransferResponse
// @Router /transfer/{transferID} [delete]
func (h *Handler) DeleteTransfer(c *gin.Context) {
	ctx := c.Request.Context()

	session, ok := request.CurrentSession(ctx)
	if !ok {
		response.Unauthorized(c.Writer, fmt.Errorf("user not authorized"))
		return
	}

	transferID, err := uuid.Parse(c.Param("transferID"))
	if err != nil {
		response.BadRequest(c.Writer, fmt.Errorf("invalid transferID: %w", err))
		return
	}

	result, err := h.transferService.DeleteTransfer(ctx, transfer.DeleteTransferInput{
		UserID:     session.UserID,
		TransferID: transferID,
	})
	if err != nil {
		response.FromError(c.Writer, err)
		return
	}

	res := DeleteTransferResponse{
		Message: result.Message,
	}

	response.OK(c.Writer, res)
}
