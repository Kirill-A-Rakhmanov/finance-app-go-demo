package httputil_test

import (
	"testing"

	"finance-app-go-demo/internal/adapter/http/httputil"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseOptionalUUID(t *testing.T) {
	id := uuid.New()
	valid, empty, invalid := id.String(), "", "not-a-uuid"

	tests := []struct {
		name    string
		raw     *string
		want    *uuid.UUID
		wantErr bool
	}{
		{name: "nil", raw: nil},
		{name: "empty string", raw: &empty},
		{name: "valid", raw: &valid, want: &id},
		{name: "invalid", raw: &invalid, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := httputil.ParseOptionalUUID(tt.raw)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseUUIDs(t *testing.T) {
	first, second := uuid.New(), uuid.New()

	tests := []struct {
		name    string
		raw     []string
		want    []uuid.UUID
		wantErr bool
	}{
		{name: "nil gives empty non-nil slice", raw: nil, want: []uuid.UUID{}},
		{name: "valid ids keep their order", raw: []string{first.String(), second.String()}, want: []uuid.UUID{first, second}},
		{name: "empty strings are skipped", raw: []string{"", first.String()}, want: []uuid.UUID{first}},
		{name: "invalid id", raw: []string{first.String(), "not-a-uuid"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := httputil.ParseUUIDs(tt.raw)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMillisToTime(t *testing.T) {
	assert.Nil(t, httputil.MillisToTime(nil))

	ms := int64(1700000000000)
	got := httputil.MillisToTime(&ms)

	require.NotNil(t, got)
	assert.Equal(t, ms, got.UnixMilli())
}
