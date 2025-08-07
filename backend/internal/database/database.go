package internal

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx"
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
	conn, err := pgx.Connect(pgx.ConnConfig{
		Host:     "localhost",
		Port:     5432,
		Database: "postgres",
		User:     "postgres",
		Password: "postgres",
	})
	if err != nil {
		return err
	}
	defer conn.Close()

	DB_NAME := os.Getenv("DB_NAME")

	_, err = conn.Exec(fmt.Sprintf("CREATE DATABASE %s", DB_NAME))
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
                    user_id UUID NOT NULL,
                    token_type TEXT NOT NULL,
                    token TEXT UNIQUE NOT NULL,
                    expires_at TIMESTAMP,
                    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
                );`,
		},
		{
			name: "config table",
			sql: `
                CREATE TABLE IF NOT EXISTS config (
                    id INTEGER PRIMARY KEY DEFAULT 1,
                    mailing_active BOOLEAN NOT NULL DEFAULT false,
                    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
                );`,
		},
		// Duststicks tables
		{
			name: "character table",
			sql: `
                CREATE TABLE IF NOT EXISTS characters (
                    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                    user_id UUID UNIQUE NOT NULL,
                    name TEXT UNIQUE NOT NULL,
                    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
                );`,
		},
		{
			name: "character inventory table",
			sql: `
                CREATE TABLE IF NOT EXISTS inventories (
                    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                    character_id UUID UNIQUE NOT NULL,
                    slots INTEGER NOT NULL DEFAULT 12,
                    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
                );`,
		},
		{
			name: "inventory rows",
			sql: `
                CREATE TABLE IF NOT EXISTS inventory_rows (
                    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                    inventory_id UUID NOT NULL,
                    item_id UUID NOT NULL,
                    item_type TEXT NOT NULL,
                    quantity INTEGER NOT NULL,
                    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                    UNIQUE (inventory_id, item_id)
                );`,
		},
		{
			name: "utility items",
			sql: `
                CREATE TABLE IF NOT EXISTS utility_items (
                    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
                    name TEXT NOT NULL,
                    value INTEGER NOT NULL,
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
