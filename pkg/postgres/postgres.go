package postgres

import (
	"context"
	"finance-app-go-demo/ent"
	"finance-app-go-demo/pkg/logger"
	"fmt"
	"net"
	"net/url"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/golang-migrate/migrate/v4"
)

type Config struct {
	Host     string `envconfig:"DATABASE_HOST" required:"true" default:"localhost"`
	Port     string `envconfig:"DATABASE_PORT" required:"true" default:"5100"`
	User     string `envconfig:"DATABASE_USER" required:"true" default:"postgres"`
	Password string `envconfig:"DATABASE_PASSWORD" required:"true"`
	DBName   string `envconfig:"DATABASE_NAME" required:"true" default:"finance-app"`
}

func (c Config) URL() string {
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(c.User, c.Password),
		Host:     net.JoinHostPort(c.Host, c.Port),
		Path:     "/" + c.DBName,
		RawQuery: "sslmode=disable",
	}

	return u.String()
}

func New(c Config, log logger.Logger, logQueries bool) (*ent.Client, error) {
	databaseURL := c.URL()

	sqlDriver, err := entsql.Open(dialect.Postgres, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("database connection error: %w", err)
	}

	configurePool(sqlDriver.DB())

	var driver dialect.Driver = sqlDriver
	if logQueries {
		driver = dialect.DebugWithContext(sqlDriver, func(ctx context.Context, args ...any) {
			log.DebugContext(ctx, "sql query", "query", fmt.Sprint(args...))
		})
	}

	client := ent.NewClient(ent.Driver(driver))

	log.Info("database connection established")

	m, err := migrate.New("file://migrations", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("migration init error: %w", err)
	}
	defer func() {
		if sourceErr, databaseErr := m.Close(); sourceErr != nil || databaseErr != nil {
			log.Warn("failed to close the migrator", "sourceError", sourceErr, "databaseError", databaseErr)
		}
	}()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return nil, fmt.Errorf("migration up error: %w", err)
	}

	log.Info("migrations applied")

	return client, nil
}
