package auth

import (
	"finance-app-go-demo/internal/adapter/http/cookie"
	"finance-app-go-demo/internal/adapter/http/response"
	"finance-app-go-demo/internal/application/auth"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type VerifyResponse struct {
	UserID uuid.UUID `json:"userId" validate:"required"`
}

// @Summary Verify
// @Description Verify access token
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} VerifyResponse
// @Router /auth/verify [get]
func (h *Handler) Verify(c *gin.Context) {
	ctx := c.Request.Context()

	acc, err := c.Cookie(cookie.AccessTokenName)
	if err != nil || acc == "" {
		response.Unauthorized(c.Writer, fmt.Errorf("no access token"))
		return
	}

	result, err := h.authService.VerifyAuth(ctx, auth.VerifyInput{AccessToken: acc})
	if err != nil {
		response.Unauthorized(c.Writer, err)
		return
	}

	res := VerifyResponse{
		UserID: result.UserID,
	}

	w := c.Writer

	w.Header().Set("X-User-ID", res.UserID.String())
	response.OK(w, res)
}
