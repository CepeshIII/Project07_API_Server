package main

import (
	"context"
	"expvar"
	"log"
	"runtime"
	"runtime/pprof"

	"github.com/CepeshIII/Project07_API_Server/internal/auth"
	config "github.com/CepeshIII/Project07_API_Server/internal/config"
	"github.com/CepeshIII/Project07_API_Server/internal/db"
	"github.com/CepeshIII/Project07_API_Server/internal/logger"
	"github.com/CepeshIII/Project07_API_Server/internal/mailer"
	"github.com/CepeshIII/Project07_API_Server/internal/ratelimiter"
	"github.com/CepeshIII/Project07_API_Server/internal/store"
	"github.com/CepeshIII/Project07_API_Server/internal/store/cache"
	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"

	"github.com/joho/godotenv"
)

const version = "1.1.0"

//	@title	GopherSocial API

//	@description	API for GopherSocial, a social network for gophers
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath					/v1
//
// @securityDefinitions.apikey	ApiKeyAuth
// @in							header
// @name						Authorization
// @description.
func main() {
	// load variables from a .env file.
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on system environment variables")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// logger
	logger := logger.NewLogger(cfg.LoggerEnv)
	defer func() {
		_ = logger.Sync()
	}()

	// database
	database, err := db.New(
		cfg.DB.Address,
		cfg.DB.MaxOpenConns,
		cfg.DB.MaxIdleConns,
		cfg.DB.MaxIdleTime,
	)
	if err != nil {
		logger.Fatal(err)
	}

	defer func() {
		_ = database.Close()
	}()

	logger.Info("database connection pool established\n")

	store := store.NewStorage(database)

	// Cache
	var rdb *redis.Client

	if cfg.Redis.Enabled {
		logger.Info("Trying to setup redis cache server\n")

		rdb = cache.NewRedisClient(cfg.Redis.Addr, cfg.Redis.Pw, cfg.Redis.DB)

		err := cache.CheckConnection(context.Background(), rdb)
		if err != nil {
			logger.Errorf("faild to create Redis client: %v\n", err)
		} else {
			logger.Info("redis cache connection pool established\n")
		}

		expvar.Publish("cacheServer", expvar.Func(func() any {
			return rdb.PoolStats()
		}))
		rdb.CSCStats()
	} else {
		expvar.Publish("cacheServer", expvar.Func(func() any {
			return "disabled"
		}))
	}

	cacheStore := cache.NewRedisStore(rdb)

	mailer := mailer.NewSendgrid(
		cfg.Mail.SendGrid.APIKey,
		cfg.Mail.FromEmail,
		cfg.Env != "production",
	)

	// mailer, err := mailer.NewMailtrap(
	// 	cfg.Mail.Mailtrap.APIKey,
	// 	cfg.Mail.FromEmail,
	// 	cfg.Env != "production",
	// )

	// if err != nil {
	// 	logger.Fatal(err)
	// }

	authenticator := auth.NewJWTAuthenticator(
		cfg.Auth.JWT.Secret,
		cfg.Auth.JWT.Iss,
		cfg.Auth.JWT.Iss,
	)

	ratelimiter := ratelimiter.NewFixedWindowRateLimiter(cfg.RateLimiter.RequestsPerTimeFrame, cfg.RateLimiter.TimeFrame)
	// Initialize validator
	validate := validator.New(validator.WithRequiredStructEnabled())

	app := &application{
		config:       cfg,
		store:        store,
		logger:       logger,
		mailer:       mailer,
		auth:         authenticator,
		cacheStorage: cacheStore,
		rateLimiter:  ratelimiter,
		validator:    validate,
	}

	expvar.NewString("version").Set(version)
	expvar.Publish("database", expvar.Func(func() any {
		return database.Stats()
	}))
	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))
	expvar.Publish("numThreads", expvar.Func(func() any {
		return pprof.Lookup("threadcreate").Count()
	}))

	mux := app.mount()

	if err := app.run(mux); err != nil {
		logger.Fatal(err)
	}
}
