package report

import (
	"finance-app-go-demo/internal/adapter/http/dto"
	"finance-app-go-demo/internal/adapter/http/httputil"
	"finance-app-go-demo/internal/adapter/http/mapper"
	"finance-app-go-demo/internal/adapter/http/request"
	"finance-app-go-demo/internal/adapter/http/response"
	appreport "finance-app-go-demo/internal/application/report"
	"finance-app-go-demo/internal/domain/enum"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type GetCounterpartyReportQuery struct {
	Type            enum.TransactionType `form:"type" validate:"required" enums:"income,expense"`
	Currency        string               `form:"currency" validate:"required"`
	AccountIDs      []string             `form:"accountIds" format:"uuid" collectionFormat:"multi"`
	PeriodStart     *int64               `form:"periodStart"`
	PeriodEnd       *int64               `form:"periodEnd"`
	CounterpartyIDs []string             `form:"counterpartyIds" format:"uuid" collectionFormat:"multi"`
}

type CounterpartyReportItem struct {
	CounterpartyID *uuid.UUID `json:"counterpartyId"`
	Amount         float64    `json:"amount" validate:"required"`
	Share          float64    `json:"share" validate:"required"`
}

type GetCounterpartyReportResponse struct {
	Total          float64                  `json:"total" validate:"required"`
	Items          []CounterpartyReportItem `json:"items" validate:"required"`
	Counterparties []dto.Counterparty       `json:"counterparties" validate:"required"`
}

// @Summary Get counterparty report
// @Description Amounts of confirmed transactions grouped by counterparty in a single currency, with each counterparty's share of the total. Transfers are excluded; transactions without a counterparty are returned as an item with a null counterparty.
// @Tags report
// @Accept json
// @Produce json
// @Param request query GetCounterpartyReportQuery true "Get counterparty report query"
// @Success 200 {object} GetCounterpartyReportResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 422 {object} response.ValidationErrorResponse
// @Router /report/counterparties [get]
func (h *Handler) GetCounterpartyReport(c *gin.Context) {
	ctx := c.Request.Context()

	session, ok := request.CurrentSession(ctx)
	if !ok {
		response.Unauthorized(c.Writer, fmt.Errorf("user not authorized"))
		return
	}

	var query GetCounterpartyReportQuery

	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c.Writer, err)
		return
	}

	accountIDs, err := httputil.ParseUUIDs(query.AccountIDs)
	if err != nil {
		response.BadRequest(c.Writer, fmt.Errorf("invalid accountIds: %w", err))
		return
	}

	counterpartyIDs, err := httputil.ParseUUIDs(query.CounterpartyIDs)
	if err != nil {
		response.BadRequest(c.Writer, fmt.Errorf("invalid counterpartyIds: %w", err))
		return
	}

	result, err := h.reportService.GetCounterpartyReport(ctx, appreport.GetCounterpartyReportInput{
		UserID:          session.UserID,
		TransactionType: query.Type,
		Currency:        query.Currency,
		PeriodStart:     httputil.MillisToTime(query.PeriodStart),
		PeriodEnd:       httputil.MillisToTime(query.PeriodEnd),
		CounterpartyIDs: counterpartyIDs,
		AccountIDs:      accountIDs,
	})
	if err != nil {
		response.FromError(c.Writer, err)
		return
	}

	items := make([]CounterpartyReportItem, len(result.Items))
	counterparties := make([]dto.Counterparty, 0, len(result.Items))

	for i, item := range result.Items {
		items[i] = CounterpartyReportItem{
			Amount: item.Amount.InexactFloat64(),
			Share:  item.Share.InexactFloat64(),
		}

		if item.Counterparty != nil {
			items[i].CounterpartyID = &item.Counterparty.ID
			counterparties = append(counterparties, *mapper.ToCounterpartyDto(item.Counterparty))
		}
	}

	response.OK(c.Writer, GetCounterpartyReportResponse{
		Total:          result.Total.InexactFloat64(),
		Items:          items,
		Counterparties: counterparties,
	})
}
