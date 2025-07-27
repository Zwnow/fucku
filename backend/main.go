package main

import (
	"log"

	"fucku/config"
	"github.com/joho/godotenv"
)

func setupApp() error {
	err := godotenv.Load()
	if err != nil {
		return err
	}
	err = config.SetupDatabase()
	if err != nil {
		return err
	}
	return nil
}

func main() {
	err := setupApp()
	if err != nil {
		log.Fatalf("App setup failed: %v", err)
	}
	log.Println("Hello world")
}
