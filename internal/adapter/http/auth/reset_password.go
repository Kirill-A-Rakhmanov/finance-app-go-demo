package auth

import (
	"finance-app-go-demo/internal/adapter/http/response"
	"finance-app-go-demo/internal/application/auth"
	"fmt"

	"github.com/gin-gonic/gin"
)

type ResetPasswordRequest struct {
	Code        string `json:"code" example:"6f6b6b6c-1a2b-4c3d-9e8f-123456789abc" validate:"required"`
	NewPassword string `json:"newPassword" example:"12345678" validate:"required"`
}

type ResetPasswordResponse struct {
	Message string `json:"message" validate:"required"`
}

// @Summary Reset password
// @Description Resets the password using a recovery code
// @Tags auth
// @Accept json
// @Produce json
// @Param request body ResetPasswordRequest true "Reset password request"
// @Success 200 {object} ResetPasswordResponse
// @Router /auth/reset-password [post]
func (h *Handler) ResetPassword(c *gin.Context) {
	ctx := c.Request.Context()

	var req ResetPasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c.Writer, fmt.Errorf("invalid request body: %v", err))
		return
	}

	result, err := h.authService.ResetPassword(ctx, auth.ResetPasswordInput{
		Code:        req.Code,
		NewPassword: req.NewPassword,
	})
	if err != nil {
		response.FromError(c.Writer, err)
		return
	}

	res := ResetPasswordResponse{
		Message: result.Message,
	}

	response.OK(c.Writer, res)
}
