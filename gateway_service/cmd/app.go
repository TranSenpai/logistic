package main

import (
	"context"
	"gateway_service/internal/di"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type App struct {
	engine *gin.Engine
	server *http.Server
}

func NewApp() (*App, error) {
	ginEngine := gin.Default()

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://localhost:3000", "http://localhost:8080"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	corsConfig.ExposeHeaders = []string{"Content-Length"}
	corsConfig.AllowCredentials = true
	corsConfig.MaxAge = 12 * time.Hour
	ginEngine.Use(cors.New(corsConfig))

	err := di.Injection(ginEngine)
	if err != nil {
		return nil, err
	}

	return &App{engine: ginEngine}, nil
}

func (a *App) Start() error {
	port := os.Getenv("PORT")
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
	log.Println("Gateway Service stopped.")
}
