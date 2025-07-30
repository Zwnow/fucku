package internal

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	DBPool *pgxpool.Pool
}

func NewDatabase(connString string) (*Database, error) {
	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize db pool: %w", err)
	}

	return &Database{DBPool: pool}, nil
}

func SetupDatabase(db *Database) error {
	DB_NAME := os.Getenv("DB_NAME")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	_, err := db.DBPool.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s", DB_NAME))
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			cancel()
			return err
		}
	}
	defer cancel()

	err = SetupTables(db)
	if err != nil {
		return err
	}

	return nil
}

func SetupTables(db *Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	_, err := db.DBPool.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`)
	if err != nil {
		log.Fatalf("failed to create uuid-ossp extension: %v", err)
	}

	_, err = db.DBPool.Exec(ctx, `
        CREATE TABLE IF NOT EXISTS users (
        id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
        email TEXT UNIQUE NOT NULL,
        username TEXT UNIQUE NOT NULL,
        password TEXT NOT NULL,
        verified INTEGER NOT NULL DEFAULT 0,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )`)
	if err != nil {
		log.Fatalf("failed to create users table: %v", err)
	}
	defer cancel()

	return nil
}
