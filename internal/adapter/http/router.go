package http

import (
	"finance-app-go-demo/internal/adapter/http/account"
	"finance-app-go-demo/internal/adapter/http/auth"
	"finance-app-go-demo/internal/adapter/http/counterparty"
	"finance-app-go-demo/internal/adapter/http/currency"
	"finance-app-go-demo/internal/adapter/http/feature_flag"
	"finance-app-go-demo/internal/adapter/http/health"
	"finance-app-go-demo/internal/adapter/http/middleware"
	"finance-app-go-demo/internal/adapter/http/report"
	"finance-app-go-demo/internal/adapter/http/transaction"
	"finance-app-go-demo/internal/adapter/http/transaction_category"
	"finance-app-go-demo/internal/adapter/http/transfer"
	"finance-app-go-demo/internal/adapter/http/user"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis_rate/v10"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewRouter(
	authMiddleware *middleware.AuthMiddleware,
	rateLimitMiddleware *middleware.RateLimitMiddleware,
	loggingMiddleware *middleware.LoggingMiddleware,
	corsMiddleware *middleware.CORSMiddleware,
	authHandler *auth.Handler,
	accountHandler *account.Handler,
	transactionHandler *transaction.Handler,
	userHandler *user.Handler,
	transferHandler *transfer.Handler,
	currencyHandler *currency.Handler,
	transactionCategoryHandler *transaction_category.Handler,
	counterpartyHandler *counterparty.Handler,
	reportHandler *report.Handler,
	featureFlagHandler *feature_flag.Handler,
	healthHandler *health.Handler,
	trustedProxies []string,
	apiRateLimit redis_rate.Limit,
	swaggerEnabled bool,
) (http.Handler, error) {
	r, err := newEngine(trustedProxies)
	if err != nil {
		return nil, err
	}

	r.Use(middleware.RequestID(), corsMiddleware.Handler(), loggingMiddleware.Handler(), loggingMiddleware.Recovery())

	r.GET("/health", healthHandler.Health)

	api := r.Group("/api")
	api.Use(rateLimitMiddleware.Handler("api", apiRateLimit))

	public := api.Group("/")
	{
		auth := public.Group("/auth")
		{
			auth.POST("/register", rateLimitMiddleware.Handler("register", redis_rate.PerHour(5)), authHandler.Register)
			auth.POST("/verify-email", rateLimitMiddleware.Handler("verify-email", redis_rate.PerHour(20)), authHandler.VerifyEmail)
			auth.POST("/login", rateLimitMiddleware.Handler("login", redis_rate.PerMinute(10)), authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshTokens)
			auth.POST("/recovery", rateLimitMiddleware.Handler("recovery", redis_rate.PerHour(5)), authHandler.Recovery)
			auth.POST("/reset-password", rateLimitMiddleware.Handler("reset-password", redis_rate.PerHour(10)), authHandler.ResetPassword)
		}

		public.GET("/feature-flag/global", featureFlagHandler.GetAllFeatureFlags)

		if swaggerEnabled {
			api.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		}
	}

	protected := api.Group("/")
	protected.Use(authMiddleware.Handler())
	{
		auth := protected.Group("/auth")
		{
			auth.POST("/logout", authHandler.Logout)
			auth.GET("/verify", authHandler.Verify)
		}

		account := protected.Group("/account")
		{
			account.POST("", accountHandler.CreateAccount)
			account.PATCH("/:accountID", accountHandler.UpdateAccount)
			account.GET("/list", accountHandler.GetAccountsByUserID)
			account.GET("/:accountID", accountHandler.GetAccountByID)
			account.GET("/:accountID/balance", accountHandler.GetBalanceByAccountID)
		}

		transaction := protected.Group("/transaction")
		{
			transaction.POST("/income", transactionHandler.CreateIncomeTransaction)
			transaction.POST("/expense", transactionHandler.CreateExpenseTransaction)
			transaction.PATCH("/:transactionID", transactionHandler.UpdateTransaction)
			transaction.DELETE("/:transactionID", transactionHandler.DeleteTransaction)
			transaction.GET("/:transactionID", transactionHandler.GetTransactionByID)
			transaction.GET("/list", transactionHandler.GetTransactions)
		}

		user := protected.Group("/user")
		{
			user.GET("/current", userHandler.GetCurrentUser)
			user.PATCH("/current", userHandler.UpdateUser)
			user.GET("/current/balance", userHandler.GetCurrentUserBalance)
			user.PATCH("/current/password", userHandler.ChangePassword)
		}

		transfer := protected.Group("/transfer")
		{
			transfer.POST("", transferHandler.CreateTransfer)
			transfer.DELETE("/:transferID", transferHandler.DeleteTransfer)
		}

		currency := protected.Group("/currency")
		{
			currency.GET("/list", currencyHandler.GetCurrencies)
		}

		transactionCategory := protected.Group("/transaction-category")
		{
			transactionCategory.POST("", transactionCategoryHandler.CreateTransactionCategory)
			transactionCategory.PATCH("/:transactionCategoryID", transactionCategoryHandler.UpdateTransactionCategory)
			transactionCategory.GET("/list", transactionCategoryHandler.GetTransactionCategories)
			transactionCategory.DELETE("/:transactionCategoryID", transactionCategoryHandler.DeleteTransactionCategory)
		}

		counterparty := protected.Group("/counterparty")
		{
			counterparty.POST("", counterpartyHandler.CreateCounterparty)
			counterparty.PATCH("/:counterpartyID", counterpartyHandler.UpdateCounterparty)
			counterparty.GET("/list", counterpartyHandler.GetCounterparties)
			counterparty.DELETE("/:counterpartyID", counterpartyHandler.DeleteCounterparty)
		}

		featureFlag := protected.Group("/feature-flag")
		{
			featureFlag.GET("/list", featureFlagHandler.GetFeatureFlags)
		}

		report := protected.Group("/report")
		{
			report.GET("/categories", reportHandler.GetCategoryReport)
			report.GET("/counterparties", reportHandler.GetCounterpartyReport)
			report.GET("/income-expense", reportHandler.GetIncomeExpenseReport)
			report.GET("/balance", reportHandler.GetBalanceReport)
			report.GET("/category-dynamic", reportHandler.GetCategoryDynamicReport)
		}
	}

	return r, nil
}

func newEngine(trustedProxies []string) (*gin.Engine, error) {
	proxies := make([]string, 0, len(trustedProxies))
	for _, proxy := range trustedProxies {
		if proxy = strings.TrimSpace(proxy); proxy != "" {
			proxies = append(proxies, proxy)
		}
	}

	r := gin.New()
	if err := r.SetTrustedProxies(proxies); err != nil {
		return nil, fmt.Errorf("invalid HTTP_TRUSTED_PROXIES: %w", err)
	}

	return r, nil
}
