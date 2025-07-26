package main

import (
	"log"

	"fucku/config"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Failed to load environment: %v", err)
	}
	err = config.SetupDatabase()
	if err != nil {
		log.Fatalf("Error setting up database: %v", err)
	}
}

func main() {
	log.Println("Hello world")
}
