package app

import (
	"context"
	"finance-app-go-demo/config"
	"finance-app-go-demo/internal/adapter/http"
	httpAccount "finance-app-go-demo/internal/adapter/http/account"
	httpAuth "finance-app-go-demo/internal/adapter/http/auth"
	"finance-app-go-demo/internal/adapter/http/cookie"
	httpCounterparty "finance-app-go-demo/internal/adapter/http/counterparty"
	httpCurrency "finance-app-go-demo/internal/adapter/http/currency"
	httpFeatureFlag "finance-app-go-demo/internal/adapter/http/feature_flag"
	httpHealth "finance-app-go-demo/internal/adapter/http/health"
	"finance-app-go-demo/internal/adapter/http/middleware"
	httpReport "finance-app-go-demo/internal/adapter/http/report"
	httpTransaction "finance-app-go-demo/internal/adapter/http/transaction"
	httpTransactionCategory "finance-app-go-demo/internal/adapter/http/transaction_category"
	httpTransfer "finance-app-go-demo/internal/adapter/http/transfer"
	httpUser "finance-app-go-demo/internal/adapter/http/user"
	appAccount "finance-app-go-demo/internal/application/account"
	appAuth "finance-app-go-demo/internal/application/auth"
	appCounterparty "finance-app-go-demo/internal/application/counterparty"
	appFeatureFlag "finance-app-go-demo/internal/application/feature_flag"
	appReport "finance-app-go-demo/internal/application/report"
	appTransaction "finance-app-go-demo/internal/application/transaction"
	appTransactionCategory "finance-app-go-demo/internal/application/transaction_category"
	appTransfer "finance-app-go-demo/internal/application/transfer"
	appUser "finance-app-go-demo/internal/application/user"
	"finance-app-go-demo/internal/infrastructure/notifier"
	"finance-app-go-demo/internal/infrastructure/password"
	"finance-app-go-demo/internal/infrastructure/postgres"
	"finance-app-go-demo/internal/infrastructure/redis"
	"finance-app-go-demo/internal/infrastructure/token"
	"finance-app-go-demo/pkg/httpserver"
	"finance-app-go-demo/pkg/logger"
	"finance-app-go-demo/pkg/mailer"
	pgConfig "finance-app-go-demo/pkg/postgres"
	redisConfig "finance-app-go-demo/pkg/redis"
	"finance-app-go-demo/pkg/transaction"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis_rate/v10"
)

func Run(c config.Config) error {
	logOutput, closeLogOutput := logger.NewOutput(os.Stdout, logger.FileConfig{
		Path:       c.App.LogFile,
		MaxSizeMB:  c.App.LogMaxSizeMB,
		MaxBackups: c.App.LogMaxBackups,
		MaxAgeDays: c.App.LogMaxAgeDays,
	})
	defer func() { _ = closeLogOutput() }()

	slogLogger := slog.New(logger.NewContextHandler(slog.NewJSONHandler(logOutput, &slog.HandlerOptions{
		Level: parseLogLevel(c.App.LogLevel),
	})))
	log := logger.New(slogLogger)

	gin.SetMode(gin.ReleaseMode)

	location, err := time.LoadLocation(c.App.Timezone)
	if err != nil || location == time.Local {
		log.Error("Invalid APP_TIMEZONE, an IANA name such as UTC or Europe/Moscow is expected", "timezone", c.App.Timezone)
		return fmt.Errorf("invalid APP_TIMEZONE %q", c.App.Timezone)
	}

	client, err := pgConfig.New(c.Postgres, log, slogLogger.Enabled(context.Background(), slog.LevelDebug))
	if err != nil {
		log.Error("Failed to connect to postgres", "error", err)
		return fmt.Errorf("pgConfig.New: %w", err)
	}

	redisClient, err := redisConfig.New(c.Redis)
	if err != nil {
		log.Error("Failed to connect to redis", "error", err)
		return fmt.Errorf("redisConfig.New: %w", err)
	}

	tokenService := token.NewTokenService(
		c.App.JWTSecret,
		c.App.AccessTokenTTL,
		c.App.RefreshTokenTTL,
	)
	passwordService := password.NewPasswordService(c.App.BcryptCost)
	cookieCfg := cookie.Config{
		Secure:          c.App.CookieSecure,
		AccessTokenTTL:  c.App.AccessTokenTTL,
		RefreshTokenTTL: c.App.RefreshTokenTTL,
	}
	sessionStorage := redis.NewSessionStorage(redisClient, c.App.RefreshTokenTTL)
	recoveryStorage := redis.NewRecoveryStorage(redisClient, c.App.RecoveryCodeTTL)
	registrationStorage := redis.NewRegistrationStorage(redisClient, c.App.RegistrationTTL)
	failureLimiter := redis.NewFailureLimiter(redisClient)

	var (
		recoveryNotifier     appAuth.RecoveryNotifier
		registrationNotifier appAuth.RegistrationNotifier
	)
	if c.SMTP.Enabled() {
		smtpMailer := mailer.New(c.SMTP)
		recoveryNotifier = notifier.NewSMTPRecoveryNotifier(smtpMailer, log, c.App.FrontendResetPasswordURL, c.App.RecoveryCodeTTL)
		registrationNotifier = notifier.NewSMTPRegistrationNotifier(smtpMailer, log, c.App.FrontendVerifyEmailURL, c.App.RegistrationTTL)
		log.Info("Emails are sent through SMTP", "host", c.SMTP.Host, "port", c.SMTP.Port, "from", c.SMTP.From)
	} else {
		recoveryNotifier = notifier.NewLogRecoveryNotifier(log, c.App.FrontendResetPasswordURL)
		registrationNotifier = notifier.NewLogRegistrationNotifier(log, c.App.FrontendVerifyEmailURL)
		log.Warn("SMTP_HOST is empty: recovery codes and registration links are written to the log instead of being emailed")
	}

	txManager := transaction.NewTransactionManager(client)
	userRepo := postgres.NewUserRepository(client)
	transactionRepo := postgres.NewTransactionRepository(client)
	transactionCategoryRepo := postgres.NewTransactionCategoryRepository(client)
	counterpartyRepo := postgres.NewCounterpartyRepository(client)
	accountRepo := postgres.NewAccountRepository(client)
	transferRepo := postgres.NewTransferRepository(client)
	featureFlagRepo := postgres.NewFeatureFlagRepository(client)
	featureFlagOverrideRepo := postgres.NewFeatureFlagOverrideRepository(client)

	transactionQueryService := postgres.NewTransactionQueryService(client)
	reportQueryService := postgres.NewReportQueryService(client)

	authService := appAuth.NewService(
		userRepo,
		transactionCategoryRepo,
		featureFlagRepo,
		tokenService,
		passwordService,
		txManager,
		sessionStorage,
		recoveryStorage,
		recoveryNotifier,
		registrationStorage,
		registrationNotifier,
		failureLimiter,
	)
	accountService := appAccount.NewService(
		txManager,
		accountRepo,
		transactionRepo,
		userRepo,
	)
	transactionService := appTransaction.NewService(
		transactionRepo,
		accountRepo,
		transactionCategoryRepo,
		counterpartyRepo,
		txManager,
		transactionQueryService,
	)
	userService := appUser.NewService(
		userRepo,
		transactionRepo,
		passwordService,
		sessionStorage,
		failureLimiter,
	)
	transferService := appTransfer.NewService(
		transferRepo,
		transactionRepo,
		accountRepo,
		txManager,
	)
	transactionCategoryService := appTransactionCategory.New(
		transactionCategoryRepo,
	)
	counterpartyService := appCounterparty.New(
		counterpartyRepo,
	)

	featureFlagService := appFeatureFlag.NewService(
		featureFlagRepo,
		featureFlagOverrideRepo,
	)

	createdFlags, err := featureFlagService.EnsureDefaultFeatureFlags(context.Background())
	if err != nil {
		log.Error("Failed to seed feature flags", "error", err)

		return fmt.Errorf("featureFlagService.EnsureDefaultFeatureFlags: %w", err)
	}
	if len(createdFlags) > 0 {
		log.Info("Default feature flags created", "keys", createdFlags)
	}

	reportService := appReport.NewService(
		accountRepo,
		transactionCategoryRepo,
		counterpartyRepo,
		reportQueryService,
		location,
	)

	authMiddleware := middleware.NewAuthMiddleware(authService)
	rateLimitMiddleware := middleware.NewRateLimitMiddleware(redisClient)
	loggingMiddleware := middleware.NewLoggingMiddleware(log)
	if c.App.APIRateLimitPerMinute <= 0 {
		log.Error("Invalid RATE_LIMIT_API_PER_MINUTE, a positive number is expected", "value", c.App.APIRateLimitPerMinute)
		return fmt.Errorf("invalid RATE_LIMIT_API_PER_MINUTE %d", c.App.APIRateLimitPerMinute)
	}
	corsMiddleware, err := middleware.NewCORSMiddleware(c.App.CORSAllowedOrigins)
	if err != nil {
		log.Error("Invalid CORS_ALLOWED_ORIGINS", "error", err)
		return fmt.Errorf("middleware.NewCORSMiddleware: %w", err)
	}
	authHandler := httpAuth.NewHandler(authService, cookieCfg)
	accountHandler := httpAccount.NewHandler(accountService)
	transactionHandler := httpTransaction.NewHandler(transactionService)
	userHandler := httpUser.NewHandler(userService)
	transferHandler := httpTransfer.NewHandler(transferService)
	currencyHandler := httpCurrency.NewHandler()
	transactionCategoryHandler := httpTransactionCategory.NewHandler(transactionCategoryService)
	counterpartyHandler := httpCounterparty.NewHandler(counterpartyService)
	reportHandler := httpReport.NewHandler(reportService)
	featureFlagHandler := httpFeatureFlag.NewHandler(featureFlagService)
	healthHandler := httpHealth.NewHandler(client, redisClient)

	router, err := http.NewRouter(
		authMiddleware,
		rateLimitMiddleware,
		loggingMiddleware,
		corsMiddleware,
		authHandler,
		accountHandler,
		transactionHandler,
		userHandler,
		transferHandler,
		currencyHandler,
		transactionCategoryHandler,
		counterpartyHandler,
		reportHandler,
		featureFlagHandler,
		healthHandler,
		c.HTTP.TrustedProxies,
		redis_rate.PerMinute(c.App.APIRateLimitPerMinute),
		c.App.SwaggerEnabled,
	)
	if err != nil {
		log.Error("Failed to build router", "error", err)
		return fmt.Errorf("http.NewRouter: %w", err)
	}

	httpServer := httpserver.New(router, c.HTTP)

	log.Info("Starting application", "port", c.HTTP.Port)

	serverErr := make(chan error, 1)
	go func() {
		if err := httpServer.Start(); err != nil {
			serverErr <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		log.Error("Failed to start server", "error", err)
		return fmt.Errorf("httpServer.Start: %w", err)
	case sig := <-quit:
		log.Info("Received shutdown signal", "signal", sig.String())
	}

	log.Info("Shutting down gracefully...")

	if err := httpServer.Stop(); err != nil {
		log.Error("Failed to gracefully stop HTTP server", "error", err)
	}

	if err := client.Close(); err != nil {
		log.Error("Failed to close postgres connection", "error", err)
	}

	if err := redisClient.Close(); err != nil {
		log.Error("Failed to close redis connection", "error", err)
	}

	log.Info("Shutdown complete")

	return nil
}

func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
