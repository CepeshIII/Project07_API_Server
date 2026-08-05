package main

import (
	"github.com/CepeshIII/Project07_API_Server/internal/db"
	"github.com/CepeshIII/Project07_API_Server/internal/env"
	"github.com/CepeshIII/Project07_API_Server/internal/store"
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

	err = db.Seed(&store, conn)
	if err != nil {
		panic(err)
	}
}
