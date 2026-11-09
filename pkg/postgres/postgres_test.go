package postgres_test

import (
	"net/url"
	"testing"

	"finance-app-go-demo/pkg/postgres"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_URL_EscapesCredentials(t *testing.T) {
	passwords := []string{
		"plain",
		"p@ss/w:rd#1?x",
		"with space and 'quotes' \"double\"",
		`back\slash%percent`,
		"юникод",
	}

	for _, password := range passwords {
		t.Run(password, func(t *testing.T) {
			c := postgres.Config{Host: "db", Port: "5432", User: "app user", Password: password, DBName: "finance-app"}

			u, err := url.Parse(c.URL())
			require.NoError(t, err)

			gotPassword, _ := u.User.Password()
			assert.Equal(t, "postgres", u.Scheme)
			assert.Equal(t, "app user", u.User.Username())
			assert.Equal(t, password, gotPassword)
			assert.Equal(t, "db:5432", u.Host)
			assert.Equal(t, "/finance-app", u.Path)
			assert.Equal(t, "disable", u.Query().Get("sslmode"))
		})
	}
}

func TestConfig_URL_BracketsIPv6Host(t *testing.T) {
	c := postgres.Config{Host: "::1", Port: "5432", User: "u", Password: "p", DBName: "d"}

	assert.Equal(t, "postgres://u:p@[::1]:5432/d?sslmode=disable", c.URL())
}
