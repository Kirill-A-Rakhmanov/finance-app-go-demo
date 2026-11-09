package transfer_test

import (
	"context"

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

// runTx makes the mocked transaction manager actually execute the wrapped
// callback, the same way the real transactionManager.Wrap does.
func runTx(txManager *transactionmock.MockManager) {
	txManager.EXPECT().
		Wrap(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, txFunc func(context.Context) error) error {
			return txFunc(ctx)
		})
	txManager.EXPECT().Lock(mock.Anything, mock.Anything).Return(nil).Maybe()
}
