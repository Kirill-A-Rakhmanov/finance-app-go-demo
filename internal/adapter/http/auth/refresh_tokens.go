package auth

import (
	"finance-app-go-demo/internal/adapter/http/cookie"
	"finance-app-go-demo/internal/adapter/http/response"
	"finance-app-go-demo/internal/application/auth"
	"fmt"

	"github.com/gin-gonic/gin"
)

type RefreshTokensResponse struct {
	Message string `json:"message" validate:"required"`
}

// @Summary Refresh tokens
// @Description Refresh token pair and set cookies
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} RefreshTokensResponse
// @Router /auth/refresh [post]
func (h *Handler) RefreshTokens(c *gin.Context) {
	ctx := c.Request.Context()

	ref, err := c.Cookie(cookie.RefreshTokenName)
	if err != nil || ref == "" {
		response.Unauthorized(c.Writer, fmt.Errorf("no refresh token"))
		return
	}

	result, err := h.authService.RefreshTokens(ctx, auth.RefreshInput{RefreshToken: ref})
	if err != nil {
		h.cookies.ClearCookies(c.Writer)

		response.Unauthorized(c.Writer, err)
		return
	}

	tokenPair := result.TokenPair

	res := RefreshTokensResponse{
		Message: "refresh tokens successfully",
	}

	h.cookies.SetTokensToCookie(c.Writer, tokenPair.AccessToken, tokenPair.RefreshToken)

	response.OK(c.Writer, res)
}
