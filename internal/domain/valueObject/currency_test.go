package valueObject_test

import (
	"errors"
	"testing"

	domainerror "finance-app-go-demo/internal/domain/error"
	"finance-app-go-demo/internal/domain/valueObject"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCurrency(t *testing.T) {
	tests := []struct {
		name         string
		code         valueObject.CurrencyCode
		wantErr      bool
		wantSymbol   string
		wantDecimals int
	}{
		{"known currency with standard decimals", valueObject.CurrencyUSD, false, "$", 2},
		{"known currency with zero decimals", valueObject.CurrencyJPY, false, "¥", 0},
		{"known currency with three decimals", valueObject.CurrencyBHD, false, ".د.ب", 3},
		{"unknown currency code", valueObject.CurrencyCode("XXX"), true, "", 0},
		{"empty currency code", valueObject.CurrencyCode(""), true, "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			currency, err := valueObject.NewCurrency(tt.code)

			if tt.wantErr {
				require.Error(t, err)
				assert.True(t, errors.Is(err, domainerror.ErrUnknownCurrencyCode))
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.code, currency.Code)
			assert.Equal(t, tt.wantSymbol, currency.Symbol)
			assert.Equal(t, tt.wantDecimals, currency.Decimals)
		})
	}
}

func TestAllCurrencies(t *testing.T) {
	currencies := valueObject.AllCurrencies()

	assert.NotEmpty(t, currencies)
	assert.Len(t, currencies, len(valueObject.Currencies))
}
