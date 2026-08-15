package main

import (
	"log"
)

func main() {
	log.Println("Starting notification_service...")
	app := NewApp()
	if err := app.Run(); err != nil {
		log.Fatalf("Failed to run notification_service: %v", err)
	}
}
