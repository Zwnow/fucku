package internal

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DBPool *pgxpool.Pool

func InitPool(c context.Context) error {
	DB_USER := os.Getenv("DB_USER")
	DB_PASSWORD := os.Getenv("DB_PASSWORD")
	DB_PORT := os.Getenv("DB_PORT")
	DB_NAME := os.Getenv("DB_NAME")

	connString := fmt.Sprintf("postgresql://%s:%s@localhost:%s/%s",
		DB_USER, DB_PASSWORD, DB_PORT, DB_NAME,
	)

	var err error
	DBPool, err = pgxpool.New(c, connString)
	if err != nil {
		log.Fatalf("Failed to initiate connection pool: %v", err)
		os.Exit(1)
	}

	return nil
}
