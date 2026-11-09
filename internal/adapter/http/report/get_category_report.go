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

type GetCategoryReportQuery struct {
	Type                   enum.TransactionType `form:"type" validate:"required" enums:"income,expense"`
	Currency               string               `form:"currency" validate:"required"`
	AccountIDs             []string             `form:"accountIds" format:"uuid" collectionFormat:"multi"`
	PeriodStart            *int64               `form:"periodStart"`
	PeriodEnd              *int64               `form:"periodEnd"`
	TransactionCategoryIDs []string             `form:"transactionCategoryIds" format:"uuid" collectionFormat:"multi"`
	GroupSubcategories     bool                 `form:"groupSubcategories"`
}

type CategoryReportItem struct {
	TransactionCategoryID *uuid.UUID `json:"transactionCategoryId"`
	Amount                float64    `json:"amount" validate:"required"`
	Share                 float64    `json:"share" validate:"required"`
}

type GetCategoryReportResponse struct {
	Total                 float64                   `json:"total" validate:"required"`
	Items                 []CategoryReportItem      `json:"items" validate:"required"`
	TransactionCategories []dto.TransactionCategory `json:"transactionCategories" validate:"required"`
}

// @Summary Get category report
// @Description Amounts of confirmed transactions grouped by category in a single currency, with each category's share of the total. Transfers are excluded; transactions without a category are returned as an item with a null category.
// @Tags report
// @Accept json
// @Produce json
// @Param request query GetCategoryReportQuery true "Get category report query"
// @Success 200 {object} GetCategoryReportResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 422 {object} response.ValidationErrorResponse
// @Router /report/categories [get]
func (h *Handler) GetCategoryReport(c *gin.Context) {
	ctx := c.Request.Context()

	session, ok := request.CurrentSession(ctx)
	if !ok {
		response.Unauthorized(c.Writer, fmt.Errorf("user not authorized"))
		return
	}

	var query GetCategoryReportQuery

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

	result, err := h.reportService.GetCategoryReport(ctx, appreport.GetCategoryReportInput{
		UserID:                 session.UserID,
		TransactionType:        query.Type,
		Currency:               query.Currency,
		PeriodStart:            httputil.MillisToTime(query.PeriodStart),
		PeriodEnd:              httputil.MillisToTime(query.PeriodEnd),
		TransactionCategoryIDs: transactionCategoryIDs,
		GroupSubcategories:     query.GroupSubcategories,
		AccountIDs:             accountIDs,
	})
	if err != nil {
		response.FromError(c.Writer, err)
		return
	}

	items := make([]CategoryReportItem, len(result.Items))
	categories := make([]dto.TransactionCategory, 0, len(result.Items))

	for i, item := range result.Items {
		items[i] = CategoryReportItem{
			Amount: item.Amount.InexactFloat64(),
			Share:  item.Share.InexactFloat64(),
		}

		if item.TransactionCategory != nil {
			items[i].TransactionCategoryID = &item.TransactionCategory.ID
			categories = append(categories, *mapper.ToTransactionCategoryDto(item.TransactionCategory))
		}
	}

	response.OK(c.Writer, GetCategoryReportResponse{
		Total:                 result.Total.InexactFloat64(),
		Items:                 items,
		TransactionCategories: categories,
	})
}
