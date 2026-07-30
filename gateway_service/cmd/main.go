package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

// @title 				Gateway Service API
// @version 			1.0
// @description 		Gateway Service cho Logistics OS.
// @host 				localhost:8080
// @BasePath 			/api/v1
func main() {
	app, err := NewApp()
	if err != nil {
		log.Fatalf("Failed to initialize Gateway App: %v", err)
	}

	if err := app.Start(); err != nil {
		log.Fatalf("Failed to start Gateway Service: %v", err)
	}

	// Đợi signal để shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	app.Stop()
}
