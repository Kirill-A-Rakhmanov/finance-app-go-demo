package password_test

import (
	"testing"

	"finance-app-go-demo/internal/infrastructure/password"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testCost = 4

func TestPasswordService_HashAndVerify(t *testing.T) {
	svc := password.NewPasswordService(testCost)

	hashed, err := svc.Hash("Password123")
	require.NoError(t, err)

	assert.True(t, svc.Verify("Password123", hashed))
	assert.False(t, svc.Verify("Password124", hashed))
}

func TestPasswordService_VerifyWithoutHashIsAlwaysFalse(t *testing.T) {
	svc := password.NewPasswordService(testCost)

	assert.False(t, svc.Verify("Password123", ""))
	assert.False(t, svc.Verify("dummy-password", ""), "the dummy password must not log anyone in")
}

func TestPasswordService_InvalidCostDoesNotPanicOnVerify(t *testing.T) {
	svc := password.NewPasswordService(99)

	assert.False(t, svc.Verify("Password123", ""))
}
