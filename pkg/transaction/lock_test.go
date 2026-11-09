package transaction

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLockID_IsStableAndDistinguishesKeys(t *testing.T) {
	assert.Equal(t, lockID("account-balance:1"), lockID("account-balance:1"))
	assert.NotEqual(t, lockID("account-balance:1"), lockID("account-balance:2"))
	assert.NotEqual(t, lockID("account-balance:1"), lockID("user-balance:1"))
}
