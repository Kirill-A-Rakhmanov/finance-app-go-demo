//go:build integration

package postgres_test

import (
	"context"
	"finance-app-go-demo/ent"
	"finance-app-go-demo/internal/domain/enum"
	"finance-app-go-demo/internal/domain/model"
	"finance-app-go-demo/internal/domain/valueObject"
	"finance-app-go-demo/internal/infrastructure/postgres"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

var testClient *ent.Client

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx, "postgres:17",
		tcpostgres.WithDatabase("finance-app-test"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to start postgres container:", err)
		os.Exit(1)
	}

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to get connection string:", err)
		os.Exit(1)
	}

	if err := applyMigrations(connStr); err != nil {
		fmt.Fprintln(os.Stderr, "failed to apply migrations:", err)
		os.Exit(1)
	}

	client, err := ent.Open(dialect.Postgres, connStr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to open ent client:", err)
		os.Exit(1)
	}
	testClient = client

	code := m.Run()

	_ = client.Close()
	_ = container.Terminate(ctx)

	os.Exit(code)
}

func applyMigrations(databaseURL string) error {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("failed to resolve migrations path")
	}

	repoRoot := filepath.Join(filepath.Dir(thisFile), "..", "..", "..")
	sourceURL := "file://" + filepath.ToSlash(filepath.Join(repoRoot, "migrations"))

	m, err := migrate.New(sourceURL, databaseURL)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}

// dec parses a decimal literal for test fixtures.
func dec(s string) decimal.Decimal {
	return decimal.RequireFromString(s)
}

// testDate returns a UTC date at midnight, `daysAgo` days before a fixed reference day,
// so results are deterministic and survive Postgres microsecond precision.
func testDate(daysAgo int) time.Time {
	return time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC).AddDate(0, 0, -daysAgo)
}

func createTestUser(t *testing.T, ctx context.Context) *model.User {
	t.Helper()

	userRepo := postgres.NewUserRepository(testClient)

	email, err := valueObject.NewEmail(fmt.Sprintf("user-%s@example.com", uuid.NewString()))
	require.NoError(t, err)

	created, err := userRepo.Create(ctx, model.NewUser("Jane", "Doe", email, "hashed-password", nil))
	require.NoError(t, err)

	return created
}

func createTestAccount(
	t *testing.T,
	ctx context.Context,
	userID uuid.UUID,
	code valueObject.CurrencyCode,
	adjustment decimal.Decimal,
) *model.Account {
	t.Helper()

	currency, err := valueObject.NewCurrency(code)
	require.NoError(t, err)

	created, err := postgres.NewAccountRepository(testClient).Create(
		ctx,
		model.NewAccount(userID, "Account "+uuid.NewString()[:8], "icon", currency, false, false, false, adjustment),
	)
	require.NoError(t, err)

	return created
}

func createTestCategory(
	t *testing.T,
	ctx context.Context,
	userID uuid.UUID,
	type_ enum.TransactionCategoryType,
	parentID *uuid.UUID,
) *model.TransactionCategory {
	t.Helper()

	created, err := postgres.NewTransactionCategoryRepository(testClient).Create(
		ctx,
		model.NewTransactionCategory(userID, parentID, "Category "+uuid.NewString()[:8], "icon", "#ffffff", type_),
	)
	require.NoError(t, err)

	return created
}

func createTestCounterparty(t *testing.T, ctx context.Context, userID uuid.UUID) *model.Counterparty {
	t.Helper()

	created, err := postgres.NewCounterpartyRepository(testClient).Create(
		ctx,
		model.NewCounterparty(userID, "Counterparty "+uuid.NewString()[:8]),
	)
	require.NoError(t, err)

	return created
}

func createTestTransaction(
	t *testing.T,
	ctx context.Context,
	accountID uuid.UUID,
	type_ enum.TransactionType,
	amount decimal.Decimal,
	date time.Time,
	isConfirmed bool,
	categoryID *uuid.UUID,
) *model.Transaction {
	t.Helper()

	created, err := postgres.NewTransactionRepository(testClient).Create(
		ctx,
		model.NewTransaction(accountID, categoryID, nil, nil, type_, amount, nil, date, isConfirmed),
	)
	require.NoError(t, err)

	return created
}
