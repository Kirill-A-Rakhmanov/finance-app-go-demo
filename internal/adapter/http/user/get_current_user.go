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

type GetCurrentUserResponse struct {
	User *dto.User `json:"user" validate:"required"`
}

// @Summary Get current user
// @Description Get current authorized user
// @Tags user
// @Accept json
// @Produce json
// @Success 200 {object} GetCurrentUserResponse
// @Router /user/current [get]
func (h *Handler) GetCurrentUser(c *gin.Context) {
	ctx := c.Request.Context()

	session, ok := request.CurrentSession(ctx)
	if !ok {
		response.Unauthorized(c.Writer, fmt.Errorf("user not authorized"))
		return
	}

	result, err := h.userService.GetUserById(ctx, user.GetUserByIdInput{
		UserID: session.UserID,
	})
	if err != nil {
		response.FromError(c.Writer, err)
		return
	}

	res := GetCurrentUserResponse{
		User: mapper.ToUserDto(result.User),
	}

	response.OK(c.Writer, res)
}
