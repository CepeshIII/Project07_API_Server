// Package config handles application configuration loading and parsing from environment variables.
package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config holds all application configuration structures.
type Config struct {
	Addr        string `envconfig:"ADDR" default:":8080"`
	ExternalURL string `envconfig:"EXTERMAL_URL" default:"localhost:8080"`
	FrontendURL string `envconfig:"FRONTEND_URL" default:"http://localhost:4000"`
	ViteAPIURL  string `envconfig:"VITE_API_URL" default:"127.0.0.1:7777"`
	Env         string `envconfig:"ENV" default:"production"`
	LoggerEnv   string `envconfig:"LOGGER_ENV" default:"development"`
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
	WriteTimeout    time.Duration `envconfig:"SERVER_WRITE_TIMEOUT" default:"30s"`
	ReadTimeout     time.Duration `envconfig:"SERVER_READ_TIMEOUT" default:"10s"`
	IdleTimeout     time.Duration `envconfig:"SERVER_IDLE_TIMEOUT" default:"1m"`
	ShutdownTimeout time.Duration `envconfig:"SERVER_SHUTDOWN_TIMEOUT" default:"5s"`
	RequestTimeout  time.Duration `envconfig:"REQUEST_TIMEOUT" default:"60s"`
}

// DBConfig holds database-related configuration parameters.
type DBConfig struct {
	Address        string `envconfig:"DB_ADDR" default:"postgres://admin:adminpassword@localhost:5433/social?sslmode=disable"`
	MaxOpenConns   int    `envconfig:"DB_MAX_OPEN_CONNS" default:"30"`
	MaxIdleConns   int    `envconfig:"DB_MAX_IDLE_CONNS" default:"30"`
	MaxIdleTime    string `envconfig:"DB_MAX_IDLE_TIME" default:"15m"`
	MigrationsPath string `envconfig:"MIGRATIONS_PATH" default:"./cmd/migrate/migrations"`
}

// MailConfig holds mailer and email delivery configuration parameters.
type MailConfig struct {
	Exp        time.Duration `envconfig:"INVITATION_EXP" default:"24h"`
	FromEmail  string        `envconfig:"FROM_EMAIL" default:""`
	MaxRetries int           `envconfig:"MAX_RETRIES_TO_SEND_EMAIL" default:"1"`
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
	Iss    string `envconfig:"AUTH_JWT_ISS" default:"project07"`
}

// TokensConfig holds token expiration settings.
type TokensConfig struct {
	AccessTokenExp  time.Duration `envconfig:"ACCESS_TOKEN_EXP" default:"15m"`
	SessionTokenExp time.Duration `envconfig:"SESSION_TOKEN_EXP" default:"720h"`
}

// RedisConfig holds Redis caching configuration.
type RedisConfig struct {
	Addr    string `envconfig:"REDIS_ADDR" default:"localhost:6379"`
	Pw      string `envconfig:"REDIS_PW" default:""`
	DB      int    `envconfig:"REDIS_DB" default:"0"`
	Enabled bool   `envconfig:"REDIS_ENABLED" default:"false"`
}

// RateLimiterConfig holds rate limiter throttling parameters.
type RateLimiterConfig struct {
	RequestsPerTimeFrame int           `envconfig:"RATE_LIMITER_REQUESTS_PER_TIME_FRAME" default:"100"`
	TimeFrame            time.Duration `envconfig:"RATE_LIMITER_TIME_FRAME_IN_SEC" default:"60s"`
	Enabled              bool          `envconfig:"RATE_LIMITER_ENABLED" default:"false"`
}

// PaginationConfig holds pagination default settings and bounds for list endpoints.
type PaginationConfig struct {
	// DefaultLimit is the default number of items returned per page if none is specified.
	DefaultLimit int `envconfig:"PAGINATION_DEFAULT_LIMIT" default:"20"`

	// MaxLimit is the maximum allowable number of items per page to prevent heavy queries from straining the database.
	MaxLimit int `envconfig:"PAGINATION_MAX_LIMIT" default:"100"`
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
