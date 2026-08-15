package main

import (
	"context"
	"log"
	"net"
	"os"
	"time"

	"user_service/ent"
	"user_service/internal/conf"
	"user_service/internal/di"

	_ "github.com/lib/pq"
	"github.com/logistic/pkg/tracer"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

type App struct {
	client   *ent.Client
	server   *grpc.Server
	shutdown func(context.Context) error
	cfg      *conf.Config
}

func NewApp(cfg *conf.Config) *App {
	shutdownTracer, err := tracer.InitTracer("user_service", os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
	if err != nil {
		log.Printf("Failed to initialize tracer: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)

	client, err := di.Injection(grpcServer, cfg)
	if err != nil {
		log.Fatalf("DI Injection failed: %v", err)
	}

	return &App{
		client:   client,
		server:   grpcServer,
		shutdown: shutdownTracer,
		cfg:      cfg,
	}
}

func (a *App) Run() error {
	port := a.cfg.Server.GrpcPort
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("UserService is listening on %v", lis.Addr())
	return a.server.Serve(lis)
}

func (a *App) Stop() {
	log.Println("Stopping UserService gracefully...")
	if a.client != nil {
		a.client.Close()
	}
	a.server.GracefulStop()

	if a.shutdown != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := a.shutdown(ctx); err != nil {
			log.Printf("Failed to shutdown tracer: %v", err)
		}
	}
	log.Println("UserService stopped.")
}
