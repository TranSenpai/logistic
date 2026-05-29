package main

import (
	"goBackend/matching_service/internal/di"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

type App struct {
	engine *gin.Engine
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
	return a.engine.Run(":8081") // Cổng mặc định cho Matching Service
}
