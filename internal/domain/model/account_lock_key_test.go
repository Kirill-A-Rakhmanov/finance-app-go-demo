package model_test

import (
	"testing"

	"finance-app-go-demo/internal/domain/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAccountBalanceLockKey(t *testing.T) {
	id := uuid.New()
	other := uuid.New()

	assert.Equal(t, model.AccountBalanceLockKey(id), model.AccountBalanceLockKey(id))
	assert.NotEqual(t, model.AccountBalanceLockKey(id), model.AccountBalanceLockKey(other))
	assert.Contains(t, model.AccountBalanceLockKey(id), id.String())
}
