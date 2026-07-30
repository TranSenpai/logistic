package main

import (
	"auth_service/internal/conf"
	"auth_service/internal/di"
	"log"
	"net"

	"google.golang.org/grpc"
)

type App struct {
	grpcServer *grpc.Server
	cfg        *conf.Config
}

func NewApp(cfg *conf.Config) (*App, error) {
	grpcServer := grpc.NewServer()

	err := di.Injection(grpcServer, cfg)
	if err != nil {
		return nil, err
	}

	return &App{grpcServer: grpcServer, cfg: cfg}, nil
}

func (a *App) Start() error {
	listener, err := net.Listen("tcp", ":"+a.cfg.Server.GrpcPort)
	if err != nil {
		return err
	}

	log.Printf("Starting Auth Service (gRPC) on :%s...", a.cfg.Server.GrpcPort)

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
