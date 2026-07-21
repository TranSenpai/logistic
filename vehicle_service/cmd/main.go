package main

import (
	"log"
)

func main() {
	log.Println("Starting vehicle_service...")
	app := NewApp()
	if err := app.Run(); err != nil {
		log.Fatalf("Failed to run vehicle_service: %v", err)
	}
}
