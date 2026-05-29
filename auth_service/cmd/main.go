package main

import (
	"log"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load("configs/.env"); err != nil {
		log.Println("No .env file found or failed to load, falling back to system environment variables")
	}

	app, err := NewApp()
	if err != nil {
		log.Fatalf("Failed to initialize Auth App: %v", err)
	}

	log.Println("Starting Auth Service on :8080...")
	if err := app.Start(); err != nil {
		log.Fatalf("Auth Service crashed: %v", err)
	}
}
