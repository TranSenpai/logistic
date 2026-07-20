package main

import (
	"log"
)

func main() {
	app, err := NewApp()
	if err != nil {
		log.Fatalf("Failed to initialize Matching App: %v", err)
	}

	log.Println("Starting Matching Service on :8081...")
	if err := app.Start(); err != nil {
		log.Fatalf("Matching Service crashed: %v", err)
	}
}
