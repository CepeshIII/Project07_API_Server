package main

import (
	"expvar"
	"log"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/CepeshIII/Project07_API_Server/internal/auth"
	"github.com/CepeshIII/Project07_API_Server/internal/db"
	"github.com/CepeshIII/Project07_API_Server/internal/env"
	"github.com/CepeshIII/Project07_API_Server/internal/logger"
	"github.com/CepeshIII/Project07_API_Server/internal/mailer"
	"github.com/CepeshIII/Project07_API_Server/internal/ratelimiter"
	"github.com/CepeshIII/Project07_API_Server/internal/store"
	"github.com/CepeshIII/Project07_API_Server/internal/store/cache"
	"github.com/redis/go-redis/v9"

	"github.com/joho/godotenv"
)

const version = "0.0.1"

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
// @description
func main() {

	// load variables from a .env file.
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	cfg := getDefaultConfig()

	// logger
	logger := logger.NewLogger(cfg.loggerEnv)
	defer logger.Sync()

	// database
	database, err := db.New(
		cfg.db.address,
		cfg.db.maxOpenConns,
		cfg.db.maxIdleConns,
		cfg.db.maxIdleTime,
	)

	if err != nil {
		logger.Fatal(err)
	}

	defer database.Close()
	logger.Info("database connection pool established\n")

	store := store.NewStorage(database)

	// Cache
	var rdb *redis.Client
	if cfg.redisCfg.enabled {
		rdb = cache.NewRedisClient(cfg.redisCfg.addr, cfg.redisCfg.pw, cfg.redisCfg.db)
		logger.Info("redis cache connection pool established\n")
	}

	cacheStore := cache.NewRedisStore(rdb)

	mailer := mailer.NewSendgrid(
		cfg.mail.sendGrid.apikey,
		cfg.mail.fromEmail,
		cfg.env != "production",
	)
	// mailer, err := mailer.NewMailtrap(
	// 	cfg.mail.mailtrap.apikey,
	// 	cfg.mail.fromEmail,
	// 	cfg.env != "production",
	// )

	// if err != nil {
	// 	logger.Fatal(err)
	// }

	authenticator := auth.NewJWTAuthenticator(
		cfg.auth.jwtAuth.jwtSecret,
		cfg.auth.jwtAuth.iss,
		cfg.auth.jwtAuth.iss,
	)

	ratelimiter := ratelimiter.NewFixedWindowRateLimiter(cfg.rateLimiter.RequestsPerTimeFrame, cfg.rateLimiter.TimeFrame)

	app := &application{
		config:       cfg,
		store:        store,
		logger:       logger,
		mailer:       mailer,
		auth:         authenticator,
		cacheStorage: cacheStore,
		rateLimiter:  ratelimiter,
	}

	expvar.NewString("version").Set(version)
	expvar.Publish("database", expvar.Func(func() any {
		return database.Stats()
	}))
	expvar.Publish("cacheServer", expvar.Func(func() any {
		return rdb.PoolStats()
	}))
	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))
	expvar.Publish("numThreads", expvar.Func(func() any {
		return pprof.Lookup("threadcreate").Count()
	}))

	rdb.CSCStats()

	mux := app.mount()

	if err := app.run(mux); err != nil {
		logger.Fatal(err)
	}

}

func getDefaultConfig() config {
	return config{
		addr:        env.GetString("ADDR", ":8080"),
		apiURL:      env.GetString("EXTERMAL_URL", "localhost:8080"),
		frontendURL: env.GetString("FRONTEND_URL", "http://localhost:4000"),
		db:          DefaultDBConfig(),
		env:         env.GetString("ENV", "production"),
		loggerEnv:   env.GetString("LOGGER_ENV", "production"),
		mail: mailConfig{
			exp:        time.Duration(env.GetInt("INVITATION_EXP", 24)) * time.Hour,
			fromEmail:  env.GetString("FROM_EMAIL", ""),
			maxRetries: env.GetInt("MAX_RETRIES_TO_SEND_EMAIL", 1),

			sendGrid: sendGridConfig{
				apikey: env.GetString("SENDGRID_API_KEY", ""),
			},
			mailtrap: mailtrapConfig{
				apikey: env.GetString("MAILTRAP_API_KEY", ""),
			},
		},
		auth: authConfig{
			basic: basicAuthConfig{
				username: env.GetString("AUTH_BASIC_USERNAME", "admin"),
				password: env.GetString("AUTH_BASIC_PASSWORD", "admin"),
			},
			jwtAuth: jwtAuthConfig{
				jwtSecret: env.GetString("AUTH_JWT_SECRET", "your-super-secret-key-from-env"),
				iss:       "project07",
			},
			tokens: tokensConfig{
				accessTokenExp:  time.Duration(env.GetInt("ACCESS_TOKEN_EXP", 15)) * time.Minute,
				sessionTokenExp: time.Duration(env.GetInt("SESSION_TOKEN_EXP", 24*30)) * time.Hour,
			},
		},
		redisCfg: redisConfig{
			addr:    env.GetString("REDIS_ADDR", "localhost:6379"),
			pw:      env.GetString("REDIS_PW", ""),
			db:      env.GetInt("REDIS_DB", 0),
			enabled: env.GetBoolean("REDIS_ENABLED", false),
		},
		rateLimiter: ratelimiter.Config{
			RequestsPerTimeFrame: env.GetInt("RATE_LIMITER_REQUESTS_PER_TIME_FRAME", 100),
			TimeFrame:            time.Duration(env.GetInt("RATE_LIMITER_TIME_FRAME_IN_SEC", 60)) * time.Second,
			Enabled:              env.GetBoolean("RATE_LIMITER_ENABLED", false),
		},
	}
}
