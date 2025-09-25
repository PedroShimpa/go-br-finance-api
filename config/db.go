package config

import (
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var DB *sqlx.DB

func ConnectDB() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/gofinance?sslmode=disable"
	}

	var err error
	DB, err = sqlx.Connect("postgres", dbURL)
	if err != nil {
		log.Fatal("❌ Erro ao conectar no banco:", err)
	}

	log.Println("✅ Conectado ao PostgreSQL")
}
