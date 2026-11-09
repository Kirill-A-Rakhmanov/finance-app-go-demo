package transaction_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"finance-app-go-demo/internal/adapter/http/request"
	httptransaction "finance-app-go-demo/internal/adapter/http/transaction"
	"finance-app-go-demo/internal/application/transaction"
	"finance-app-go-demo/internal/domain/enum"
	"finance-app-go-demo/internal/domain/model"
	repomock "finance-app-go-demo/internal/domain/repository/mock"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// The JSON fields transactionCategoryName / counterpartyName must reach the service and end up as
// ids on the stored transaction; a name that does not exist yet is created on the fly.
func TestHandler_TransactionNamesAreResolvedToIDs(t *testing.T) {
	session := &request.Session{UserID: uuid.New()}
	accountID := uuid.New()
	transactionID := uuid.New()
	categoryID := uuid.New()
	counterpartyID := uuid.New()

	tests := []struct {
		name         string
		method       string
		path         string
		categoryType enum.TransactionCategoryType
		body         string
		setupStored  func(transactionRepo *repomock.MockTransactionRepository, account *model.Account)
		mount        func(r *gin.Engine, h *httptransaction.Handler)
	}{
		{
			name:         "create expense",
			method:       http.MethodPost,
			path:         "/transaction/expense",
			categoryType: enum.Expense,
			body:         `{"accountId":"` + accountID.String() + `","amount":50,"transactionCategoryName":"Groceries","counterpartyName":"Supermarket"}`,
			setupStored: func(transactionRepo *repomock.MockTransactionRepository, account *model.Account) {
				account.IsNegativeBalanceEnabled = true
				transactionRepo.EXPECT().Create(mock.Anything, mock.Anything).
					RunAndReturn(func(ctx context.Context, tx *model.Transaction) (*model.Transaction, error) {
						tx.ID = transactionID
						return tx, nil
					})
			},
			mount: func(r *gin.Engine, h *httptransaction.Handler) {
				r.POST("/transaction/expense", h.CreateExpenseTransaction)
			},
		},
		{
			name:         "create income",
			method:       http.MethodPost,
			path:         "/transaction/income",
			categoryType: enum.Income,
			body:         `{"accountId":"` + accountID.String() + `","amount":50,"transactionCategoryName":"Groceries","counterpartyName":"Supermarket"}`,
			setupStored: func(transactionRepo *repomock.MockTransactionRepository, account *model.Account) {
				transactionRepo.EXPECT().Create(mock.Anything, mock.Anything).
					RunAndReturn(func(ctx context.Context, tx *model.Transaction) (*model.Transaction, error) {
						tx.ID = transactionID
						return tx, nil
					})
			},
			mount: func(r *gin.Engine, h *httptransaction.Handler) {
				r.POST("/transaction/income", h.CreateIncomeTransaction)
			},
		},
		{
			name:         "update",
			method:       http.MethodPut,
			path:         "/transaction/" + transactionID.String(),
			categoryType: enum.Income,
			body: `{"accountId":"` + accountID.String() + `","amount":50,"type":"income","isConfirmed":true,"date":"2026-06-30T00:00:00Z",` +
				`"transactionCategoryName":"Groceries","counterpartyName":"Supermarket"}`,
			setupStored: func(transactionRepo *repomock.MockTransactionRepository, account *model.Account) {
				transactionRepo.EXPECT().GetByIDAndUserID(mock.Anything, transactionID, session.UserID).
					Return(&model.Transaction{ID: transactionID, AccountID: accountID, Type: enum.TransactionIncome}, nil)
				transactionRepo.EXPECT().Update(mock.Anything, mock.Anything).
					RunAndReturn(func(ctx context.Context, tx *model.Transaction) (*model.Transaction, error) {
						return tx, nil
					})
			},
			mount: func(r *gin.Engine, h *httptransaction.Handler) {
				r.PUT("/transaction/:transactionID", h.UpdateTransaction)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accountRepo := repomock.NewMockAccountRepository(t)
			transactionRepo := repomock.NewMockTransactionRepository(t)
			categoryRepo := repomock.NewMockTransactionCategoryRepository(t)
			counterpartyRepo := repomock.NewMockCounterpartyRepository(t)

			account := &model.Account{ID: accountID, UserID: session.UserID}
			tt.setupStored(transactionRepo, account)
			accountRepo.EXPECT().GetByIDAndUserID(mock.Anything, accountID, session.UserID).Return(account, nil)

			categoryRepo.EXPECT().GetByUserIDAndTypeAndName(mock.Anything, session.UserID, tt.categoryType, "Groceries").Return(nil, nil)
			categoryRepo.EXPECT().Create(mock.Anything, mock.Anything).
				RunAndReturn(func(ctx context.Context, c *model.TransactionCategory) (*model.TransactionCategory, error) {
					assert.Equal(t, "Groceries", c.Name)
					assert.Equal(t, tt.categoryType, c.Type)
					c.ID = categoryID
					return c, nil
				})
			counterpartyRepo.EXPECT().GetByUserIDAndName(mock.Anything, session.UserID, "Supermarket").Return(nil, nil)
			counterpartyRepo.EXPECT().Create(mock.Anything, mock.Anything).
				RunAndReturn(func(ctx context.Context, c *model.Counterparty) (*model.Counterparty, error) {
					c.ID = counterpartyID
					return c, nil
				})

			service := transaction.NewService(transactionRepo, accountRepo, categoryRepo, counterpartyRepo, newPassthroughTxManager(t), nil)

			gin.SetMode(gin.TestMode)
			r := gin.New()
			r.Use(withSession(session))
			tt.mount(r, httptransaction.NewHandler(service))

			req := httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

			var body map[string]map[string]any
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			assert.Equal(t, categoryID.String(), body["transaction"]["transactionCategoryId"])
			assert.Equal(t, counterpartyID.String(), body["transaction"]["counterpartyId"])
		})
	}
}
