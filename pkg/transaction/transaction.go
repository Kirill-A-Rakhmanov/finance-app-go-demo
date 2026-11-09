package transaction

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"

	"finance-app-go-demo/ent"
)

type contextKey struct{}

var entTxKey = contextKey{}

var ErrNoActiveTransaction = errors.New("no active transaction")

type Manager interface {
	Wrap(ctx context.Context, txFunc func(ctx context.Context) error) error
	Lock(ctx context.Context, key string) error
}

type transactionManager struct {
	client *ent.Client
}

func NewTransactionManager(client *ent.Client) Manager {
	return &transactionManager{client: client}
}

func (t *transactionManager) Wrap(ctx context.Context, txFunc func(context.Context) error) error {
	tx, err := t.client.Tx(ctx)
	if err != nil {
		return err
	}

	txCtx := context.WithValue(ctx, entTxKey, tx)

	err = txFunc(txCtx)

	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return err
		}

		return err
	}

	return tx.Commit()
}

func (t *transactionManager) Lock(ctx context.Context, key string) error {
	tx, ok := ctx.Value(entTxKey).(*ent.Tx)
	if !ok {
		return ErrNoActiveTransaction
	}

	_, err := tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock($1)", lockID(key))

	return err
}

func lockID(key string) int64 {
	sum := sha256.Sum256([]byte(key))

	return int64(binary.BigEndian.Uint64(sum[:8]))
}

func GetClient(ctx context.Context, defaultClient *ent.Client) *ent.Client {
	if tx, ok := ctx.Value(entTxKey).(*ent.Tx); ok {
		return tx.Client()
	}

	return defaultClient
}
