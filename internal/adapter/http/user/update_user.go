package user

import (
	"finance-app-go-demo/internal/adapter/http/dto"
	"finance-app-go-demo/internal/adapter/http/mapper"
	"finance-app-go-demo/internal/adapter/http/request"
	"finance-app-go-demo/internal/adapter/http/response"
	"finance-app-go-demo/internal/application/user"
	"fmt"

	"github.com/gin-gonic/gin"
)

type UpdateUserRequest struct {
	FirstName string `json:"firstName" example:"John" validate:"required"`
	LastName  string `json:"lastName" example:"Doe" validate:"required"`
}

type UpdateUserResponse struct {
	User dto.User `json:"user" validate:"required"`
}

// @Summary Update current user
// @Description Update the first and last name of the current user
// @Tags user
// @Accept json
// @Produce json
// @Param request body UpdateUserRequest true "Update user request"
// @Success 200 {object} UpdateUserResponse
// @Router /user/current [patch]
func (h *Handler) UpdateUser(c *gin.Context) {
	ctx := c.Request.Context()

	session, ok := request.CurrentSession(ctx)
	if !ok {
		response.Unauthorized(c.Writer, fmt.Errorf("user not authorized"))
		return
	}

	var req UpdateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c.Writer, fmt.Errorf("invalid request body: %v", err))
		return
	}

	result, err := h.userService.UpdateUser(ctx, user.UpdateUserInput{
		UserID:    session.UserID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})
	if err != nil {
		response.FromError(c.Writer, err)
		return
	}

	res := UpdateUserResponse{
		User: *mapper.ToUserDto(result.User),
	}

	response.OK(c.Writer, res)
}
