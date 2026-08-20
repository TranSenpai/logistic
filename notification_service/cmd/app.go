package main

import (
	"context"
	"errors"
	"log"
	"net"
	"os"
	"time"

	"notification_service/internal/conf"
	"notification_service/internal/di"

	_ "github.com/lib/pq"
	"github.com/logistic/pkg/middleware"
	"github.com/logistic/pkg/tracer"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

type App struct {
	container  *di.Container
	server     *grpc.Server
	shutdown   func(context.Context) error
	cfg        *conf.Config
	cancelCons context.CancelFunc
}

func NewApp(cfg *conf.Config) (*App, error) {
	shutdownTracer, err := tracer.InitTracer("notification_service", os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
	if err != nil {
		log.Printf("Failed to initialize tracer: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		middleware.ChainForService("notification_service"),
	)

	container, err := di.Injection(grpcServer, cfg)
	if err != nil {
		return nil, err
	}

	return &App{
		container: container,
		server:    grpcServer,
		shutdown:  shutdownTracer,
		cfg:       cfg,
	}, nil
}

// Run khởi động consumer RabbitMQ trong goroutine rồi block ở gRPC server.
//
// Consumer chạy nền vì Start() của nó block tới khi context bị huỷ; nếu để chung
// luồng thì gRPC server sẽ không bao giờ được Serve.
func (a *App) Run() error {
	if a.container.HasConsumer() {
		ctx, cancel := context.WithCancel(context.Background())
		a.cancelCons = cancel

		go func() {
			log.Println("NotificationService: consumer RabbitMQ bắt đầu chạy")
			if err := a.container.StartConsumer(ctx); err != nil && !errors.Is(err, context.Canceled) {
				// Consumer chết KHÔNG kéo theo cả service: API đọc inbox vẫn phải
				// sống. Nhưng phải log rõ để bên vận hành biết luồng sự kiện đã đứt.
				log.Printf("NotificationService: consumer RabbitMQ dừng bất thường: %v", err)
			}
		}()
	}

	lis, err := net.Listen("tcp", ":"+a.cfg.Server.GrpcPort)
	if err != nil {
		return err
	}
	log.Printf("NotificationService is listening on %v", lis.Addr())
	return a.server.Serve(lis)
}

func (a *App) Stop() {
	log.Println("Stopping NotificationService gracefully...")

	// Dừng consumer TRƯỚC: message đang xử lý dở được chạy nốt, còn message mới
	// thì nằm lại queue chờ instance khác — không mất mát gì.
	if a.cancelCons != nil {
		a.cancelCons()
	}

	a.server.GracefulStop()
	a.container.Close()

	if a.shutdown != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := a.shutdown(ctx); err != nil {
			log.Printf("Failed to shutdown tracer: %v", err)
		}
	}
	log.Println("NotificationService stopped.")
}
