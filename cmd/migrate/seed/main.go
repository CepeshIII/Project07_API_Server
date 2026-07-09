package main

import (
	"course/api_server/internal/db"
	"course/api_server/internal/env"
	"course/api_server/internal/store"
)

func main() {
	addr := env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost:5433/social?sslmode=disable")

	conn, err := db.New(
		addr,
		env.GetInt("DB_MAX_OPEN_CONNS", 3),
		env.GetInt("DB_MAX_IDLE_CONNS", 3),
		env.GetString("DB_MAX_IDLE_TIME", "15m"),
	)

	if err != nil {
		panic(err)
	}

	defer conn.Close()

	store := store.NewStorage(conn)

	err = db.Seed(&store)
	if err != nil {
		panic(err)
	}
}
