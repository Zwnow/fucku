package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func SetupDatabase() error {
	DB_USER := os.Getenv("DB_USER")
	DB_PASSWORD := os.Getenv("DB_PASSWORD")
	DB_PORT := os.Getenv("DB_PORT")
	DB_NAME := os.Getenv("DB_NAME")

	connStr := fmt.Sprintf("user=%s password=%s port=%s dbname=postgres sslmode=disable", DB_USER, DB_PASSWORD, DB_PORT)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", DB_NAME))
	if err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}

	return nil
}
