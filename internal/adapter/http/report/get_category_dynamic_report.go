package report

import (
	"finance-app-go-demo/internal/adapter/http/dto"
	"finance-app-go-demo/internal/adapter/http/httputil"
	"finance-app-go-demo/internal/adapter/http/mapper"
	"finance-app-go-demo/internal/adapter/http/request"
	"finance-app-go-demo/internal/adapter/http/response"
	appreport "finance-app-go-demo/internal/application/report"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type GetCategoryDynamicReportQuery struct {
	Currency               string   `form:"currency" validate:"required"`
	TransactionCategoryIDs []string `form:"transactionCategoryIds" validate:"required" format:"uuid" collectionFormat:"multi"`
	AccountIDs             []string `form:"accountIds" format:"uuid" collectionFormat:"multi"`
	PeriodStart            *int64   `form:"periodStart"`
	PeriodEnd              *int64   `form:"periodEnd"`
}

type CategoryDynamicReportItem struct {
	Month                 string    `json:"month" validate:"required" example:"2026-03"`
	TransactionCategoryID uuid.UUID `json:"transactionCategoryId" validate:"required"`
	Amount                float64   `json:"amount" validate:"required"`
}

type GetCategoryDynamicReportResponse struct {
	Months                []string                    `json:"months" validate:"required"`
	Items                 []CategoryDynamicReportItem `json:"items" validate:"required"`
	TransactionCategories []dto.TransactionCategory   `json:"transactionCategories" validate:"required"`
}

// @Summary Get category dynamic report
// @Description Amounts of confirmed transactions by calendar month and category in a single currency
// @Tags report
// @Accept json
// @Produce json
// @Param request query GetCategoryDynamicReportQuery true "Get category dynamic report query"
// @Success 200 {object} GetCategoryDynamicReportResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 422 {object} response.ValidationErrorResponse
// @Router /report/category-dynamic [get]
func (h *Handler) GetCategoryDynamicReport(c *gin.Context) {
	ctx := c.Request.Context()

	session, ok := request.CurrentSession(ctx)
	if !ok {
		response.Unauthorized(c.Writer, fmt.Errorf("user not authorized"))
		return
	}

	var query GetCategoryDynamicReportQuery

	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c.Writer, err)
		return
	}

	accountIDs, err := httputil.ParseUUIDs(query.AccountIDs)
	if err != nil {
		response.BadRequest(c.Writer, fmt.Errorf("invalid accountIds: %w", err))
		return
	}

	transactionCategoryIDs, err := httputil.ParseUUIDs(query.TransactionCategoryIDs)
	if err != nil {
		response.BadRequest(c.Writer, fmt.Errorf("invalid transactionCategoryIds: %w", err))
		return
	}

	result, err := h.reportService.GetCategoryDynamicReport(ctx, appreport.GetCategoryDynamicReportInput{
		UserID:                 session.UserID,
		Currency:               query.Currency,
		PeriodStart:            httputil.MillisToTime(query.PeriodStart),
		PeriodEnd:              httputil.MillisToTime(query.PeriodEnd),
		TransactionCategoryIDs: transactionCategoryIDs,
		AccountIDs:             accountIDs,
	})
	if err != nil {
		response.FromError(c.Writer, err)
		return
	}

	items := make([]CategoryDynamicReportItem, len(result.Items))
	for i, item := range result.Items {
		items[i] = CategoryDynamicReportItem{
			Month:                 item.Month,
			TransactionCategoryID: item.TransactionCategoryID,
			Amount:                item.Amount.InexactFloat64(),
		}
	}

	categories := make([]dto.TransactionCategory, len(result.TransactionCategories))
	for i := range result.TransactionCategories {
		categories[i] = *mapper.ToTransactionCategoryDto(&result.TransactionCategories[i])
	}

	response.OK(c.Writer, GetCategoryDynamicReportResponse{
		Months:                result.Months,
		Items:                 items,
		TransactionCategories: categories,
	})
}
