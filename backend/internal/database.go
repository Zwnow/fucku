package internal

import (
	"context"
	"fmt"
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
	defer cancel()

	_, err := db.DBPool.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s", DB_NAME))
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return err
		}
	}

	err = SetupTables(db)
	if err != nil {
		return err
	}

	return nil
}

func SetupTables(db *Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	queries := []struct {
		name string
		sql  string
	}{
		{
			name: "uuid-ossp extension",
			sql:  `CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`,
		},
		{
			name: "users table",
			sql: `
                CREATE TABLE IF NOT EXISTS users (
                    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                    email TEXT UNIQUE NOT NULL,
                    username TEXT UNIQUE NOT NULL,
                    password TEXT NOT NULL,
                    verified INTEGER NOT NULL DEFAULT 0,
                    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
                );`,
		},
		{
			name: "tokens table",
			sql: `
                CREATE TABLE IF NOT EXISTS tokens (
                    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                    token_type TEXT NOT NULL,
                    token TEXT UNIQUE NOT NULL,
                    expires_at TIMESTAMP,
                    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
                );`,
		},
	}

	for _, q := range queries {
		if _, err := db.DBPool.Exec(ctx, q.sql); err != nil {
			return fmt.Errorf("failed to create %s: %w", q.name, err)
		}
	}

	return nil
}
