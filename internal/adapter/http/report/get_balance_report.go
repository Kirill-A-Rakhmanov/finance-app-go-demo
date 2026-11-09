package report

import (
	"finance-app-go-demo/internal/adapter/http/httputil"
	"finance-app-go-demo/internal/adapter/http/request"
	"finance-app-go-demo/internal/adapter/http/response"
	appreport "finance-app-go-demo/internal/application/report"
	"fmt"

	"github.com/gin-gonic/gin"
)

type GetBalanceReportQuery struct {
	Currency    string   `form:"currency" validate:"required"`
	AccountIDs  []string `form:"accountIds" format:"uuid" collectionFormat:"multi"`
	PeriodStart *int64   `form:"periodStart"`
	PeriodEnd   *int64   `form:"periodEnd"`
}

type BalanceReportItem struct {
	DateKey string  `json:"dateKey" validate:"required" example:"2026-03-01"`
	Balance float64 `json:"balance" validate:"required"`
}

type GetBalanceReportResponse struct {
	Granularity string              `json:"granularity" validate:"required" enums:"day,week,month"`
	Items       []BalanceReportItem `json:"items" validate:"required"`
}

// @Summary Get balance report
// @Description How the total balance of the accounts changed over a period, in a single currency
// @Tags report
// @Accept json
// @Produce json
// @Param request query GetBalanceReportQuery true "Get balance report query"
// @Success 200 {object} GetBalanceReportResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 422 {object} response.ValidationErrorResponse
// @Router /report/balance [get]
func (h *Handler) GetBalanceReport(c *gin.Context) {
	ctx := c.Request.Context()

	session, ok := request.CurrentSession(ctx)
	if !ok {
		response.Unauthorized(c.Writer, fmt.Errorf("user not authorized"))
		return
	}

	var query GetBalanceReportQuery

	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c.Writer, err)
		return
	}

	accountIDs, err := httputil.ParseUUIDs(query.AccountIDs)
	if err != nil {
		response.BadRequest(c.Writer, fmt.Errorf("invalid accountIds: %w", err))
		return
	}

	result, err := h.reportService.GetBalanceReport(ctx, appreport.GetBalanceReportInput{
		UserID:      session.UserID,
		Currency:    query.Currency,
		PeriodStart: httputil.MillisToTime(query.PeriodStart),
		PeriodEnd:   httputil.MillisToTime(query.PeriodEnd),
		AccountIDs:  accountIDs,
	})
	if err != nil {
		response.FromError(c.Writer, err)
		return
	}

	items := make([]BalanceReportItem, len(result.Items))
	for i, item := range result.Items {
		items[i] = BalanceReportItem{
			DateKey: item.DateKey,
			Balance: item.Balance.InexactFloat64(),
		}
	}

	response.OK(c.Writer, GetBalanceReportResponse{
		Granularity: string(result.Granularity),
		Items:       items,
	})
}
