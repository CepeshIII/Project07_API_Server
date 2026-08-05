package main

import (
	"log"
	"time"

	"github.com/CepeshIII/Project07_API_Server/internal/db"
	"github.com/CepeshIII/Project07_API_Server/internal/env"
	"github.com/CepeshIII/Project07_API_Server/internal/logger"
	"github.com/CepeshIII/Project07_API_Server/internal/mailer"
	"github.com/CepeshIII/Project07_API_Server/internal/store"

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

	cfg := config{
		addr:        env.GetString("ADDR", ":8080"),
		apiURL:      env.GetString("EXTERMAL_URL", "localhost:8080"),
		frontendURL: env.GetString("FRONTEND_URL", "http://localhost:4000"),
		db:          DefaultDBConfig(),
		env:         env.GetString("ENV", "production"),
		mail: mailConfig{
			exp:       time.Duration(env.GetInt("INVITATION_EXP", 24)) * time.Hour,
			fromEmail: env.GetString("FROM_EMAIL", ""),
			sendGrid: sendGridConfig{
				apikey: env.GetString("SENDGRID_API_KEY", ""),
			},
		},
	}

	// logger
	logger := logger.NewLogger(cfg.env)
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
	mailer := mailer.NewSendgrid(
		cfg.mail.sendGrid.apikey,
		cfg.mail.fromEmail,
	)

	app := &application{
		config: cfg,
		store:  store,
		logger: logger,
		mailer: mailer,
	}

	mux := app.mount()

	if err := app.run(mux); err != nil {
		logger.Fatal(err)
	}

}
