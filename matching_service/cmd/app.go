package main

import (
	"matching_service/internal/conf"
	"matching_service/internal/di"
	"net"

	_ "github.com/go-sql-driver/mysql"
	"google.golang.org/grpc"
)

type App struct {
	grpcServer *grpc.Server
	listener   net.Listener
	cfg        *conf.Config
}

func NewApp(cfg *conf.Config) (*App, error) {
	lis, err := net.Listen("tcp", ":"+cfg.Server.GrpcPort)
	if err != nil {
		return nil, err
	}

	grpcServer := grpc.NewServer()

	err = di.Injection(grpcServer, cfg)
	if err != nil {
		return nil, err
	}

	return &App{
		grpcServer: grpcServer,
		listener:   lis,
		cfg:        cfg,
	}, nil
}

func (a *App) Start() error {
	return a.grpcServer.Serve(a.listener)
}

func (a *App) Stop() {
	log.Println("Stopping Matching Service (gRPC) gracefully...")
	a.grpcServer.GracefulStop()
	log.Println("Matching Service (gRPC) stopped.")
}
