package error_test

import (
	"strings"
	"testing"

	apperror "finance-app-go-demo/internal/application/error"

	"github.com/stretchr/testify/assert"
)

func TestValidationError_MaxLength(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{name: "empty", value: "", wantErr: false},
		{name: "exactly max", value: strings.Repeat("a", 5), wantErr: false},
		{name: "over max", value: strings.Repeat("a", 6), wantErr: true},
		{name: "counts characters not bytes", value: strings.Repeat("я", 5), wantErr: false},
		{name: "multibyte over max", value: strings.Repeat("я", 6), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := apperror.NewValidationError()
			v.MaxLength("name", tt.value, 5)

			assert.Equal(t, tt.wantErr, v.HasErrors())
			if tt.wantErr {
				assert.Equal(t, []string{"must be at most 5 characters"}, v.Fields["name"])
			}
		})
	}
}

func TestValidationError_MaxLengthPtr(t *testing.T) {
	long := strings.Repeat("a", 6)
	short := "abc"

	v := apperror.NewValidationError()
	v.MaxLengthPtr("nilField", nil, 5)
	v.MaxLengthPtr("shortField", &short, 5)
	assert.False(t, v.HasErrors())

	v.MaxLengthPtr("longField", &long, 5)
	assert.Contains(t, v.Fields, "longField")
	assert.Len(t, v.Fields, 1)
}
