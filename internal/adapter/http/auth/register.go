package auth

import (
	"finance-app-go-demo/internal/adapter/http/response"
	"finance-app-go-demo/internal/application/auth"
	"fmt"

	"github.com/gin-gonic/gin"
)

type RegisterRequest struct {
	FirstName string `json:"firstName" example:"John" validate:"required"`
	LastName  string `json:"lastName" example:"Doe" validate:"required"`
	Email     string `json:"email" example:"john@example.com" validate:"required"`
	Password  string `json:"password" example:"12345678" validate:"required"`
}

type RegisterResponse struct {
	Message string `json:"message" validate:"required"`
}

// @Summary Register user
// @Description Starts a registration and emails a confirmation link
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Register request"
// @Success 200 {object} RegisterResponse
// @Router /auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	ctx := c.Request.Context()

	var req RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c.Writer, fmt.Errorf("invalid request body: %v", err))
		return
	}

	result, err := h.authService.Register(ctx, auth.RegisterInput{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Password:  req.Password,
	})
	if err != nil {
		response.FromError(c.Writer, err)
		return
	}

	response.OK(c.Writer, RegisterResponse{Message: result.Message})
}
