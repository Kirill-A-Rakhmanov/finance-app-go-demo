package config

import (
	"finance-app-go-demo/pkg/httpserver"
	"finance-app-go-demo/pkg/mailer"
	"finance-app-go-demo/pkg/postgres"
	"finance-app-go-demo/pkg/redis"
	"fmt"
	"log"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type App struct {
	Name    string `envconfig:"APP_NAME" required:"true" default:"finance-app-go-demo"`
	Version string `envconfig:"APP_VERSION" required:"true" default:"0.8.0"`

	LogLevel      string `envconfig:"LOG_LEVEL" required:"true" default:"info"`
	LogFile       string `envconfig:"LOG_FILE"`
	LogMaxSizeMB  int    `envconfig:"LOG_MAX_SIZE_MB" required:"true" default:"50"`
	LogMaxBackups int    `envconfig:"LOG_MAX_BACKUPS" required:"true" default:"10"`
	LogMaxAgeDays int    `envconfig:"LOG_MAX_AGE_DAYS" required:"true" default:"30"`

	Timezone string `envconfig:"APP_TIMEZONE" required:"true" default:"UTC"`

	SwaggerEnabled bool `envconfig:"APP_SWAGGER_ENABLED" required:"true"`

	JWTSecret       string        `envconfig:"AUTH_JWT_SECRET" required:"true"`
	AccessTokenTTL  time.Duration `envconfig:"AUTH_ACCESS_TOKEN_TTL" required:"true" default:"15m"`
	RefreshTokenTTL time.Duration `envconfig:"AUTH_REFRESH_TOKEN_TTL" required:"true" default:"336h"`
	BcryptCost      int           `envconfig:"AUTH_BCRYPT_COST" required:"true" default:"12"`
	CookieSecure    bool          `envconfig:"AUTH_COOKIE_SECURE" required:"true"`
	RecoveryCodeTTL time.Duration `envconfig:"AUTH_RECOVERY_CODE_TTL" required:"true" default:"15m"`
	// RegistrationTTL is how long a registration waits for its email to be confirmed.
	RegistrationTTL time.Duration `envconfig:"AUTH_REGISTRATION_TTL" required:"true" default:"24h"`

	APIRateLimitPerMinute int `envconfig:"RATE_LIMIT_API_PER_MINUTE" required:"true" default:"300"`

	CORSAllowedOrigins []string `envconfig:"CORS_ALLOWED_ORIGINS" required:"true"`

	FrontendResetPasswordURL string `envconfig:"FRONTEND_RESET_PASSWORD_URL" required:"true"`
	FrontendVerifyEmailURL   string `envconfig:"FRONTEND_VERIFY_EMAIL_URL" required:"true"`
}

const JWTSecretMinLength = 32

func (a App) Validate() error {
	if len(a.JWTSecret) < JWTSecretMinLength {
		return fmt.Errorf("AUTH_JWT_SECRET must be at least %d characters long, got %d (generate one with: openssl rand -hex 32)", JWTSecretMinLength, len(a.JWTSecret))
	}

	return nil
}

type Config struct {
	App      App
	Postgres postgres.Config
	Redis    redis.Config
	HTTP     httpserver.Config
	SMTP     mailer.Config
}

func InitConfig() (Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	var cfg Config

	if err := envconfig.Process("", &cfg.App); err != nil {
		return Config{}, err
	}

	if err := cfg.App.Validate(); err != nil {
		return Config{}, err
	}

	if err := envconfig.Process("", &cfg.Postgres); err != nil {
		return Config{}, err
	}

	if err := envconfig.Process("", &cfg.Redis); err != nil {
		return Config{}, err
	}

	if err := envconfig.Process("", &cfg.HTTP); err != nil {
		return Config{}, err
	}

	if err := envconfig.Process("", &cfg.SMTP); err != nil {
		return Config{}, err
	}

	if err := cfg.SMTP.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
