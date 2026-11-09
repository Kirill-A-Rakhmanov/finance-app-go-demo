package counterparty

import (
	"finance-app-go-demo/internal/adapter/http/dto"
	"finance-app-go-demo/internal/adapter/http/mapper"
	"finance-app-go-demo/internal/adapter/http/request"
	"finance-app-go-demo/internal/adapter/http/response"
	"finance-app-go-demo/internal/application/counterparty"
	"fmt"

	"github.com/gin-gonic/gin"
)

type CreateCounterpartyRequest struct {
	Name string `json:"name" example:"Supermarket" validate:"required"`
}

type CreateCounterpartyResponse struct {
	Counterparty dto.Counterparty `json:"counterparty" validate:"required"`
}

// @Summary Create counterparty
// @Description Create new counterparty for current user
// @Tags counterparty
// @Accept json
// @Produce json
// @Param request body CreateCounterpartyRequest true "Create counterparty request"
// @Success 200 {object} CreateCounterpartyResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 422 {object} response.ValidationErrorResponse
// @Router /counterparty [post]
func (h *Handler) CreateCounterparty(c *gin.Context) {
	ctx := c.Request.Context()

	session, ok := request.CurrentSession(ctx)
	if !ok {
		response.Unauthorized(c.Writer, fmt.Errorf("user not authorized"))
		return
	}

	var req CreateCounterpartyRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c.Writer, fmt.Errorf("invalid request body: %v", err))
		return
	}

	result, err := h.counterpartyService.CreateCounterparty(ctx, counterparty.CreateCounterpartyInput{
		UserID: session.UserID,
		Name:   req.Name,
	})
	if err != nil {
		response.FromError(c.Writer, err)
		return
	}

	res := CreateCounterpartyResponse{
		Counterparty: *mapper.ToCounterpartyDto(result.Counterparty),
	}

	response.OK(c.Writer, res)
}
