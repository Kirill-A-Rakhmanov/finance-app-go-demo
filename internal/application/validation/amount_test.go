package validation_test

import (
	"testing"

	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/application/validation"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestAmount(t *testing.T) {
	tests := []struct {
		name        string
		amount      string
		wantMessage string
	}{
		{"typical", "1234.56", ""},
		{"smallest allowed", "0.0001", ""},
		{"whole number", "100", ""},
		{"trailing zeros do not count as decimals", "100.00000", ""},
		{"exactly the maximum", "1000000000000", ""},
		{"zero", "0", "must be greater than zero"},
		{"negative", "-5", "must be greater than zero"},
		{"above the maximum", "1000000000000.0001", "must not exceed 1000000000000"},
		{"far above the maximum", "99999999999999999999", "must not exceed 1000000000000"},
		{"too many decimals", "1.00001", "must have at most 4 decimal places"},
		{"tiny positive amount", "0.00001", "must have at most 4 decimal places"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := apperror.NewValidationError()

			validation.Amount(v, "amount", decimal.RequireFromString(tt.amount))

			if tt.wantMessage == "" {
				assert.False(t, v.HasErrors(), "%v", v.Fields)
				return
			}

			assert.Equal(t, []string{tt.wantMessage}, v.Fields["amount"])
		})
	}
}

func TestAmount_ReportsEveryProblemOfAnAmountThatIsBothTooBigAndTooPrecise(t *testing.T) {
	v := apperror.NewValidationError()

	validation.Amount(v, "amount", decimal.RequireFromString("2000000000000.00001"))

	assert.Len(t, v.Fields["amount"], 2)
}
