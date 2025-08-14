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
		{
			name: "song of the day table",
			sql: `
                CREATE TABLE IF NOT EXISTS songs (
                    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
					song_name TEXT NOT NULL,
					album_name TEXT NOT NULL,
					artist TEXT NOT NULL,
					featuring_artist TEXT DEFAULT NULL,
                    spotify_embed_url TEXT NOT NULL,
					reason TEXT NOT NULL,
                    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
                );`,
		},
		{
			name: "genre tags",
			sql: `
                CREATE TABLE IF NOT EXISTS genres (
					id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
					genre_name TEXT NOT NULL UNIQUE,
                    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
                );`,
		},
		{
			name: "special tags",
			sql: `
                CREATE TABLE IF NOT EXISTS special_tags (
					id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
					name TEXT NOT NULL UNIQUE,
					description TEXT,
                    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
                );`,
		},
		{
			name: "song genre mapping tags",
			sql: `
                CREATE TABLE IF NOT EXISTS song_genres (
					song_id UUID NOT NULL REFERENCES songs(id) ON DELETE CASCADE,
					genre_id INTEGER NOT NULL REFERENCES genres(id) ON DELETE CASCADE,
					PRIMARY KEY (song_id, genre_id)
                );`,
		},
		{
			name: "song special tag mapping",
			sql: `
                CREATE TABLE IF NOT EXISTS song_special_tags (
					song_id UUID NOT NULL REFERENCES songs(id) ON DELETE CASCADE,
					tag_id INTEGER NOT NULL REFERENCES special_tags(id) ON DELETE CASCADE,
					PRIMARY KEY (song_id, tag_id)
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
