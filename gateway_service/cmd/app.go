package main

import (
	"context"
	"gateway_service/internal/di"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/logistic/pkg/tracer"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type App struct {
	engine    *gin.Engine
	container *di.Container
	server    *http.Server
	shutdown  func(context.Context) error
}

func NewApp() (*App, error) {
	shutdownTracer, err := tracer.InitTracer("gateway_service", os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
	if err != nil {
		log.Printf("Failed to initialize tracer: %v", err)
	}

	// gin.New() thay vì gin.Default(): Logger và Recovery mặc định của gin ghi
	// log dạng text và trả về body 500 không theo khung lỗi chung của hệ thống.
	// RegisterGatewayRoutes tự gắn middleware.AccessLog + middleware.Recovery
	// để mọi phản hồi — kể cả khi panic — đều cùng một định dạng JSON.
	ginEngine := gin.New()

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://localhost:3000", "http://localhost:8080"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	corsConfig.ExposeHeaders = []string{"Content-Length"}
	corsConfig.AllowCredentials = true
	corsConfig.MaxAge = 12 * time.Hour
	ginEngine.Use(cors.New(corsConfig))

	container, err := di.Injection(ginEngine)
	if err != nil {
		return nil, err
	}

	return &App{
		engine:    ginEngine,
		container: container,
		shutdown:  shutdownTracer,
	}, nil
}

func (a *App) Start() error {
	port := os.Getenv("GATEWAY_PORT")
	if port == "" {
		log.Fatal("PORT environment variable is missing for Gateway Service")
	}

	a.server = &http.Server{
		Addr:    ":" + port,
		Handler: a.engine,
	}

	log.Printf("Starting Gateway Service on :%s...", port)

	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Gateway Service crashed: %v", err)
		}
	}()
	return nil
}

func (a *App) Stop() {
	log.Println("Stopping Gateway Service gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		log.Printf("Gateway Service forced to shutdown: %v", err)
	}

	a.container.Close()

	if a.shutdown != nil {
		if err := a.shutdown(ctx); err != nil {
			log.Printf("Failed to shutdown tracer: %v", err)
		}
	}

	log.Println("Gateway Service stopped.")
}
