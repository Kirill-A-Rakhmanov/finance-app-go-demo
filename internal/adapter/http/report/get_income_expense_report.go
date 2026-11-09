package report

import (
	"finance-app-go-demo/internal/adapter/http/httputil"
	"finance-app-go-demo/internal/adapter/http/request"
	"finance-app-go-demo/internal/adapter/http/response"
	appreport "finance-app-go-demo/internal/application/report"
	"fmt"

	"github.com/gin-gonic/gin"
)

type GetIncomeExpenseReportQuery struct {
	Currency    string   `form:"currency" validate:"required"`
	AccountIDs  []string `form:"accountIds" format:"uuid" collectionFormat:"multi"`
	PeriodStart *int64   `form:"periodStart"`
	PeriodEnd   *int64   `form:"periodEnd"`
}

type IncomeExpenseReportItem struct {
	Month   string  `json:"month" validate:"required" example:"2026-03"`
	Income  float64 `json:"income" validate:"required"`
	Expense float64 `json:"expense" validate:"required"`
}

type GetIncomeExpenseReportResponse struct {
	Items []IncomeExpenseReportItem `json:"items" validate:"required"`
}

// @Summary Get income and expense report
// @Description Income and expense of confirmed transactions by calendar month in a single currency
// @Tags report
// @Accept json
// @Produce json
// @Param request query GetIncomeExpenseReportQuery true "Get income and expense report query"
// @Success 200 {object} GetIncomeExpenseReportResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 422 {object} response.ValidationErrorResponse
// @Router /report/income-expense [get]
func (h *Handler) GetIncomeExpenseReport(c *gin.Context) {
	ctx := c.Request.Context()

	session, ok := request.CurrentSession(ctx)
	if !ok {
		response.Unauthorized(c.Writer, fmt.Errorf("user not authorized"))
		return
	}

	var query GetIncomeExpenseReportQuery

	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c.Writer, err)
		return
	}

	accountIDs, err := httputil.ParseUUIDs(query.AccountIDs)
	if err != nil {
		response.BadRequest(c.Writer, fmt.Errorf("invalid accountIds: %w", err))
		return
	}

	result, err := h.reportService.GetIncomeExpenseReport(ctx, appreport.GetIncomeExpenseReportInput{
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

	items := make([]IncomeExpenseReportItem, len(result.Items))
	for i, item := range result.Items {
		items[i] = IncomeExpenseReportItem{
			Month:   item.Month,
			Income:  item.Income.InexactFloat64(),
			Expense: item.Expense.InexactFloat64(),
		}
	}

	response.OK(c.Writer, GetIncomeExpenseReportResponse{Items: items})
}
