// Package config handles application configuration loading and parsing from environment variables.
package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config holds all application configuration structures.
type Config struct {
	Addr        string `default:":8080"                 envconfig:"ADDR"`
	ExternalURL string `default:"localhost:8080"        envconfig:"EXTERMAL_URL"`
	FrontendURL string `default:"http://localhost:4000" envconfig:"FRONTEND_URL"`
	ViteAPIURL  string `default:"127.0.0.1:7777"        envconfig:"VITE_API_URL"`
	Env         string `default:"production"            envconfig:"ENV"`
	LoggerEnv   string `default:"development"           envconfig:"LOGGER_ENV"`
	Server      ServerConfig
	DB          DBConfig
	Pagination  PaginationConfig
	Mail        MailConfig
	Auth        AuthConfig
	Redis       RedisConfig
	RateLimiter RateLimiterConfig
}

// ServerConfig holds HTTP server tuning parameters and timeouts.
type ServerConfig struct {
	WriteTimeout    time.Duration `default:"30s" envconfig:"SERVER_WRITE_TIMEOUT"`
	ReadTimeout     time.Duration `default:"10s" envconfig:"SERVER_READ_TIMEOUT"`
	IdleTimeout     time.Duration `default:"1m"  envconfig:"SERVER_IDLE_TIMEOUT"`
	ShutdownTimeout time.Duration `default:"5s"  envconfig:"SERVER_SHUTDOWN_TIMEOUT"`
	RequestTimeout  time.Duration `default:"60s" envconfig:"REQUEST_TIMEOUT"`
}

// DBConfig holds database-related configuration parameters.
type DBConfig struct {
	Address        string `default:"postgres://admin:adminpassword@localhost:5433/social?sslmode=disable" envconfig:"DB_ADDR"`
	MaxOpenConns   int    `default:"30"                                                                   envconfig:"DB_MAX_OPEN_CONNS"`
	MaxIdleConns   int    `default:"30"                                                                   envconfig:"DB_MAX_IDLE_CONNS"`
	MaxIdleTime    string `default:"15m"                                                                  envconfig:"DB_MAX_IDLE_TIME"`
	MigrationsPath string `default:"./cmd/migrate/migrations"                                             envconfig:"MIGRATIONS_PATH"`
}

// MailConfig holds mailer and email delivery configuration parameters.
type MailConfig struct {
	Exp        time.Duration `default:"24h" envconfig:"INVITATION_EXP"`
	FromEmail  string        `default:""    envconfig:"FROM_EMAIL"`
	MaxRetries int           `default:"1"   envconfig:"MAX_RETRIES_TO_SEND_EMAIL"`
	SendGrid   SendGridConfig
	Mailtrap   MailtrapConfig
}

// SendGridConfig holds SendGrid specific credentials.
type SendGridConfig struct {
	APIKey string `envconfig:"SENDGRID_API_KEY" required:"true"`
}

// MailtrapConfig holds Mailtrap specific credentials.
type MailtrapConfig struct {
	APIKey string `envconfig:"MAILTRAP_API_KEY" required:"true"`
}

// AuthConfig holds authentication and security configuration.
type AuthConfig struct {
	Basic  BasicAuthConfig
	JWT    JWTAuthConfig
	Tokens TokensConfig
}

// BasicAuthConfig holds basic auth credentials.
type BasicAuthConfig struct {
	Username string `envconfig:"AUTH_BASIC_USERNAME" required:"true"`
	Password string `envconfig:"AUTH_BASIC_PASSWORD" required:"true"`
}

// JWTAuthConfig holds JWT signing secrets and issuer information.
type JWTAuthConfig struct {
	Secret string `envconfig:"AUTH_JWT_SECRET" required:"true"`
	Iss    string `default:"project07"         envconfig:"AUTH_JWT_ISS"`
}

// TokensConfig holds token expiration settings.
type TokensConfig struct {
	AccessTokenExp  time.Duration `default:"15m"  envconfig:"ACCESS_TOKEN_EXP"`
	SessionTokenExp time.Duration `default:"720h" envconfig:"SESSION_TOKEN_EXP"`
}

// RedisConfig holds Redis caching configuration.
type RedisConfig struct {
	Addr    string `default:"localhost:6379" envconfig:"REDIS_ADDR"`
	Pw      string `default:""               envconfig:"REDIS_PW"`
	DB      int    `default:"0"              envconfig:"REDIS_DB"`
	Enabled bool   `default:"false"          envconfig:"REDIS_ENABLED"`
}

// RateLimiterConfig holds rate limiter throttling parameters.
type RateLimiterConfig struct {
	RequestsPerTimeFrame int           `default:"100"   envconfig:"RATE_LIMITER_REQUESTS_PER_TIME_FRAME"`
	TimeFrame            time.Duration `default:"60s"   envconfig:"RATE_LIMITER_TIME_FRAME_IN_SEC"`
	Enabled              bool          `default:"false" envconfig:"RATE_LIMITER_ENABLED"`
}

// PaginationConfig holds pagination default settings and bounds for list endpoints.
type PaginationConfig struct {
	// DefaultLimit is the default number of items returned per page if none is specified.
	DefaultLimit int `default:"20" envconfig:"PAGINATION_DEFAULT_LIMIT"`

	// MaxLimit is the maximum allowable number of items per page to prevent heavy queries from straining the database.
	MaxLimit int `default:"100" envconfig:"PAGINATION_MAX_LIMIT"`
}

// Load reads environment variables into the Config struct and wraps any errors for lint compliance.
func Load() (Config, error) {
	var cfg Config

	err := envconfig.Process("", &cfg)
	if err != nil {
		return Config{}, fmt.Errorf("failed to process environment configuration: %w", err)
	}

	return cfg, nil
}
