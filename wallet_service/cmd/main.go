package main

import (
	"log"
)

func main() {
	log.Println("Starting wallet_service...")
	app := NewApp()
	if err := app.Run(); err != nil {
		log.Fatalf("Failed to run wallet_service: %v", err)
	}
}
