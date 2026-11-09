package transaction_test

import (
	"context"
	"testing"

	"finance-app-go-demo/internal/adapter/http/request"
	transactionmock "finance-app-go-demo/pkg/transaction/mock"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
)

// withSession injects the given session into the request context, standing in
// for the real authMiddleware so handlers can be exercised without the rest of
// the middleware stack. A nil session leaves the context untouched, simulating
// an unauthenticated request.
func withSession(session *request.Session) gin.HandlerFunc {
	return func(c *gin.Context) {
		if session != nil {
			ctx := request.WithSession(c.Request.Context(), *session)
			c.Request = c.Request.WithContext(ctx)
		}
		c.Next()
	}
}

// newPassthroughTxManager returns a transaction manager mock that just runs the callback. Wrap is
// optional because requests rejected before the service (401, 400) never reach it.
func newPassthroughTxManager(t *testing.T) *transactionmock.MockManager {
	t.Helper()

	txManager := transactionmock.NewMockManager(t)
	txManager.EXPECT().
		Wrap(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, txFunc func(context.Context) error) error {
			return txFunc(ctx)
		}).
		Maybe()
	txManager.EXPECT().Lock(mock.Anything, mock.Anything).Return(nil).Maybe()

	return txManager
}
