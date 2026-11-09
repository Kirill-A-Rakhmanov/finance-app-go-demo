package auth

import (
	"finance-app-go-demo/internal/adapter/http/request"
	"finance-app-go-demo/internal/adapter/http/response"
	"finance-app-go-demo/internal/application/auth"
	"fmt"

	"github.com/gin-gonic/gin"
)

type LogoutResponse struct {
	Message string `json:"message" validate:"required"`
}

// @Summary Logout user
// @Description Logout existing user
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} LogoutResponse
// @Router /auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	ctx := c.Request.Context()

	session, ok := request.CurrentSession(ctx)
	if !ok {
		response.Unauthorized(c.Writer, fmt.Errorf("user not authorized"))
		return
	}

	err := h.authService.Logout(ctx, auth.LogoutInput{
		UserID:    session.UserID,
		SessionID: session.SessionID,
	})
	if err != nil {
		response.Internal(c.Writer, err)
		return
	}
	h.cookies.ClearCookies(c.Writer)

	res := LogoutResponse{
		Message: "logged out",
	}

	response.OK(c.Writer, res)
}
