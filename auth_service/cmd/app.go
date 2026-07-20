package main

import (
	"auth_service/internal/di"
	"time"

	_ "auth_service/docs"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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

	// Swagger route
	ginEngine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	err := di.Injection(ginEngine)
	if err != nil {
		return nil, err
	}

	return &App{engine: ginEngine}, nil
}

func (a *App) Start() error {
	return a.engine.Run(":8080") // Cổng mặc định cho Auth Service
}
