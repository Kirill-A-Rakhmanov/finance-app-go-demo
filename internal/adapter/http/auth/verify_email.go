package auth

import (
	"finance-app-go-demo/internal/adapter/http/response"
	"finance-app-go-demo/internal/application/auth"
	"fmt"

	"github.com/gin-gonic/gin"
)

type VerifyEmailRequest struct {
	Token string `json:"token" example:"3q2-7wQ5..." validate:"required"`
}

type VerifyEmailResponse struct {
	Message string `json:"message" validate:"required"`
}

// @Summary Confirm registration
// @Description Creates the account for the token from the confirmation email
// @Tags auth
// @Accept json
// @Produce json
// @Param request body VerifyEmailRequest true "Confirmation token"
// @Success 200 {object} VerifyEmailResponse
// @Router /auth/verify-email [post]
func (h *Handler) VerifyEmail(c *gin.Context) {
	ctx := c.Request.Context()

	var req VerifyEmailRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c.Writer, fmt.Errorf("invalid request body: %v", err))
		return
	}

	result, err := h.authService.ConfirmRegistration(ctx, auth.ConfirmRegistrationInput{
		Token: req.Token,
	})
	if err != nil {
		response.FromError(c.Writer, err)
		return
	}

	response.OK(c.Writer, VerifyEmailResponse{Message: result.Message})
}
