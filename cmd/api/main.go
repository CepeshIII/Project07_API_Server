package main

import (
	"context"
	"course/api_server/internal/db"
	"course/api_server/internal/env"
	"course/api_server/internal/store"
	"log"
	"log/slog"

	"github.com/joho/godotenv"
)

const version = "0.0.1"

func main() {
	logger := slog.Default()
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	cfg := config{
		addr: env.GetString("ADDR", ":8080"),
		db:   DefaultDBConfig(),
		env:  env.GetString("ENV", "development"),
	}

	database, err := db.New(
		cfg.db.address,
		cfg.db.maxOpenConns,
		cfg.db.maxIdleConns,
		cfg.db.maxIdleTime,
	)

	if err != nil {
		log.Panic(err)
	}

	defer database.Close()
	logger.Log(context.Background(), slog.LevelInfo, "database connection pool established\n")

	store := store.NewStorage(database)

	app := &application{
		config: cfg,
		logger: logger,
		store:  store,
	}

	mux := app.mount()

	if err := app.run(mux); err != nil {
		log.Fatalln(err)
	}

}
