package main

import (
	"auth_service/internal/di"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
)

type App struct {
	grpcServer *grpc.Server
}

func NewApp() (*App, error) {
	grpcServer := grpc.NewServer()

	err := di.Injection(grpcServer)
	if err != nil {
		return nil, err
	}

	return &App{grpcServer: grpcServer}, nil
}

func (a *App) Start() error {
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		log.Fatal("GRPC_PORT environment variable is missing")
	}

	listener, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		return err
	}
	
	log.Printf("Starting Auth Service (gRPC) on :%s...", grpcPort)
	
	// Start gRPC server in a separate goroutine
	go func() {
		if err := a.grpcServer.Serve(listener); err != nil {
			log.Fatalf("Auth Service crashed: %v", err)
		}
	}()
	return nil
}

func (a *App) Stop() {
	log.Println("Stopping Auth Service (gRPC) gracefully...")
	a.grpcServer.GracefulStop()
	log.Println("Auth Service (gRPC) stopped.")
}
