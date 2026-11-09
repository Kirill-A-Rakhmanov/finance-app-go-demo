package error_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	apperror "finance-app-go-demo/internal/application/error"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTooManyRequests(t *testing.T) {
	err := apperror.TooManyRequests("slow down", 90*time.Second)

	assert.Equal(t, apperror.KindTooManyRequests, err.Kind)
	assert.Equal(t, "slow down", err.Error())
	assert.Equal(t, 90*time.Second, err.RetryAfter)
}

func TestTooManyRequests_SurvivesWrapping(t *testing.T) {
	wrapped := fmt.Errorf("login: %w", apperror.TooManyRequests("slow down", time.Minute))

	var appErr *apperror.Error
	require.True(t, errors.As(wrapped, &appErr))
	assert.Equal(t, time.Minute, appErr.RetryAfter)
}
