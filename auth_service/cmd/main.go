package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

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

	if err := app.Start(); err != nil {
		log.Fatalf("Failed to start Auth Service: %v", err)
	}

	// Đợi signal để shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	app.Stop()
}
