package auth

import (
	"finance-app-go-demo/internal/adapter/http/dto"
	"finance-app-go-demo/internal/adapter/http/mapper"
	"finance-app-go-demo/internal/adapter/http/response"
	"finance-app-go-demo/internal/application/auth"
	"fmt"

	"github.com/gin-gonic/gin"
)

type LoginRequest struct {
	Email    string `json:"email" example:"john@example.com" validate:"required"`
	Password string `json:"password" example:"12345678" validate:"required"`
}

type LoginResponse struct {
	User dto.User `json:"user" validate:"required"`
}

// @Summary Login user
// @Description Login existing user. Unknown email and wrong password both return 401 invalid credentials. After 5 failed attempts for an email within 15 minutes the endpoint answers 429 with a Retry-After header.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login request"
// @Success 200 {object} LoginResponse
// @Router /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	ctx := c.Request.Context()

	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c.Writer, fmt.Errorf("invalid request body: %v", err))
		return
	}

	result, err := h.authService.Login(ctx, auth.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})

	if err != nil {
		response.FromError(c.Writer, err)
		return
	}

	res := LoginResponse{
		User: *mapper.ToUserDto(result.User),
	}

	h.cookies.SetTokensToCookie(c.Writer, result.AccessToken, result.RefreshToken)

	response.OK(c.Writer, res)
}
