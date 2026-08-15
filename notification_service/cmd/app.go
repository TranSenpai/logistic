package main

import "log"

type App struct {
	// Declare app dependencies here
}

func NewApp() *App {
	return &App{}
}

func (a *App) Run() error {
	log.Println("notification_service is running!")
	// Initialize HTTP or gRPC server here
	return nil
}
