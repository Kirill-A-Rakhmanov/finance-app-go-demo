package transaction

import (
	"finance-app-go-demo/internal/adapter/http/dto"
	"finance-app-go-demo/internal/adapter/http/httputil"
	"finance-app-go-demo/internal/adapter/http/mapper"
	"finance-app-go-demo/internal/adapter/http/request"
	"finance-app-go-demo/internal/adapter/http/response"
	"finance-app-go-demo/internal/application/query"
	"finance-app-go-demo/internal/application/transaction"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type GetTransactionsQuery struct {
	AccountID              *string                  `form:"accountId" format:"uuid"`
	TransactionCategoryIDs []string                 `form:"transactionCategoryIds" format:"uuid" collectionFormat:"multi"`
	PeriodStart            *int64                   `form:"periodStart"`
	PeriodEnd              *int64                   `form:"periodEnd"`
	Limit                  int                      `form:"limit" validate:"required"`
	Offset                 int                      `form:"offset" validate:"required"`
	SortBy                 *query.TransactionSortBy `form:"sortBy"`
	SortDirection          *query.SortDirection     `form:"sortDirection"`
}

type GetTransactionsResponse struct {
	Total        int               `json:"total" validate:"required"`
	Limit        int               `json:"limit" validate:"required"`
	Offset       int               `json:"offset" validate:"required"`
	Order        []uuid.UUID       `json:"order" validate:"required"`
	Transactions []dto.Transaction `json:"transactions" validate:"required"`
}

// @Summary Get transactions
// @Description Get transactions with filters, sorting and pagination
// @Tags transaction
// @Accept json
// @Produce json
// @Param request query GetTransactionsQuery true "Get transactions query"
// @Success 200 {object} GetTransactionsResponse
// @Router /transaction/list [get]
func (h *Handler) GetTransactions(c *gin.Context) {
	ctx := c.Request.Context()

	session, ok := request.CurrentSession(ctx)
	if !ok {
		response.Unauthorized(c.Writer, fmt.Errorf("user not authorized"))
		return
	}

	var query GetTransactionsQuery

	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c.Writer, err)
		return
	}

	accountID, err := httputil.ParseOptionalUUID(query.AccountID)
	if err != nil {
		response.BadRequest(c.Writer, fmt.Errorf("invalid accountId: %w", err))
		return
	}

	transactionCategoryIDs, err := httputil.ParseUUIDs(query.TransactionCategoryIDs)
	if err != nil {
		response.BadRequest(c.Writer, fmt.Errorf("invalid transactionCategoryIds: %w", err))
		return
	}

	result, err := h.transactionService.GetTransactions(ctx, transaction.GetTransactionsInput{
		UserID:                 session.UserID,
		AccountID:              accountID,
		TransactionCategoryIDs: transactionCategoryIDs,
		PeriodStart:            httputil.MillisToTime(query.PeriodStart),
		PeriodEnd:              httputil.MillisToTime(query.PeriodEnd),
		Limit:                  query.Limit,
		Offset:                 query.Offset,
		SortBy:                 query.SortBy,
		SortDirection:          query.SortDirection,
	})
	if err != nil {
		response.FromError(c.Writer, err)
		return
	}

	transactions := make([]dto.Transaction, len(result.Transactions))
	for i, transaction := range result.Transactions {
		transactions[i] = *mapper.ToTransactionDto(&transaction)
	}

	res := GetTransactionsResponse{
		Total:        result.Total,
		Limit:        result.Limit,
		Offset:       result.Offset,
		Order:        result.Order,
		Transactions: transactions,
	}

	response.OK(c.Writer, res)
}
