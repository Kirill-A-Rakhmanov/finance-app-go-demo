package auth

import (
	"finance-app-go-demo/internal/adapter/http/response"
	"finance-app-go-demo/internal/application/auth"
	"fmt"

	"github.com/gin-gonic/gin"
)

type RecoveryRequest struct {
	Email string `json:"email" example:"john@example.com" validate:"required"`
}

type RecoveryResponse struct {
	Message string `json:"message" validate:"required"`
}

// @Summary Request password recovery
// @Description Sends a password recovery code to the user's email. The response is the same whether or not the account exists, and a repeated request for the same email within 2 minutes sends nothing.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RecoveryRequest true "Recovery request"
// @Success 200 {object} RecoveryResponse
// @Router /auth/recovery [post]
func (h *Handler) Recovery(c *gin.Context) {
	ctx := c.Request.Context()

	var req RecoveryRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c.Writer, fmt.Errorf("invalid request body: %v", err))
		return
	}

	result, err := h.authService.Recovery(ctx, auth.RecoveryInput{
		Email: req.Email,
	})
	if err != nil {
		response.FromError(c.Writer, err)
		return
	}

	res := RecoveryResponse{
		Message: result.Message,
	}

	response.OK(c.Writer, res)
}
