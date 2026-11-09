package currency

import (
	"finance-app-go-demo/internal/adapter/http/dto"
	"finance-app-go-demo/internal/adapter/http/mapper"
	"finance-app-go-demo/internal/adapter/http/response"
	"finance-app-go-demo/internal/domain/valueObject"
	"sort"

	"github.com/gin-gonic/gin"
)

type GetCurrenciesResponse struct {
	Currencies []dto.Currency `json:"currencies" validate:"required"`
}

// @Summary Get currencies
// @Description Get list of all supported currencies
// @Tags currency
// @Accept json
// @Produce json
// @Success 200 {object} GetCurrenciesResponse
// @Router /currency/list [get]
func (h *Handler) GetCurrencies(c *gin.Context) {
	currencies := valueObject.AllCurrencies()

	sort.Slice(currencies, func(i, j int) bool {
		return currencies[i].Code < currencies[j].Code
	})

	result := make([]dto.Currency, len(currencies))
	for i, currency := range currencies {
		result[i] = mapper.ToCurrencyDto(currency)
	}

	response.OK(c.Writer, GetCurrenciesResponse{
		Currencies: result,
	})
}
