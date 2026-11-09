package user

import (
	"finance-app-go-demo/internal/adapter/http/request"
	"finance-app-go-demo/internal/adapter/http/response"
	"finance-app-go-demo/internal/application/user"
	"fmt"

	"github.com/gin-gonic/gin"
)

type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" example:"12345678" validate:"required"`
	NewPassword string `json:"newPassword" example:"87654321" validate:"required"`
}

type ChangePasswordResponse struct {
	Message string `json:"message" validate:"required"`
}

// @Summary Change password
// @Description Change the password of the current user and sign out all other sessions. After 5 wrong current passwords within 15 minutes the endpoint answers 429 with a Retry-After header.
// @Tags user
// @Accept json
// @Produce json
// @Param request body ChangePasswordRequest true "Change password request"
// @Success 200 {object} ChangePasswordResponse
// @Router /user/current/password [patch]
func (h *Handler) ChangePassword(c *gin.Context) {
	ctx := c.Request.Context()

	session, ok := request.CurrentSession(ctx)
	if !ok {
		response.Unauthorized(c.Writer, fmt.Errorf("user not authorized"))
		return
	}

	var req ChangePasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c.Writer, fmt.Errorf("invalid request body: %v", err))
		return
	}

	result, err := h.userService.ChangePassword(ctx, user.ChangePasswordInput{
		UserID:      session.UserID,
		SessionID:   session.SessionID,
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
	})
	if err != nil {
		response.FromError(c.Writer, err)
		return
	}

	res := ChangePasswordResponse{
		Message: result.Message,
	}

	response.OK(c.Writer, res)
}
