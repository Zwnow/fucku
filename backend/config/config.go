package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/jackc/pgx"
)

func SetupDatabase() error {
	DB_USER := os.Getenv("DB_USER")
	DB_PASSWORD := os.Getenv("DB_PASSWORD")
	DB_PORT, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		log.Fatalf("Failed to parse port from environment: %v", err)
	}
	DB_NAME := os.Getenv("DB_NAME")

	config := pgx.ConnConfig{
		User:     DB_USER,
		Password: DB_PASSWORD,
		Database: "postgres",
		Port:     uint16(DB_PORT),
		Host:     "localhost",
	}
	conn, err := pgx.Connect(config)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Exec(fmt.Sprintf("CREATE DATABASE %s", DB_NAME))
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return err
		}
	}

	err = SetupTables()
	if err != nil {
		return err
	}

	return nil
}

func SetupTables() error {
	DB_USER := os.Getenv("DB_USER")
	DB_PASSWORD := os.Getenv("DB_PASSWORD")
	DB_PORT, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		log.Fatalf("Failed to parse port from environment: %v", err)
	}
	DB_NAME := os.Getenv("DB_NAME")

	config := pgx.ConnConfig{
		User:     DB_USER,
		Password: DB_PASSWORD,
		Database: DB_NAME,
		Port:     uint16(DB_PORT),
		Host:     "localhost",
	}
	conn, err := pgx.Connect(config)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`)
	if err != nil {
		log.Fatalf("failed to create uuid-ossp extension: %v", err)
	}

	_, err = conn.Exec(`
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

	return nil
}
