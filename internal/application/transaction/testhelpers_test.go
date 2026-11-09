package transaction_test

import (
	"context"
	"testing"

	transactionmock "finance-app-go-demo/pkg/transaction/mock"

	"github.com/stretchr/testify/mock"
)

// newPassthroughTxManager returns a transaction manager mock that just runs the callback. Wrap is
// optional because use-cases that fail validation never reach it.
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
