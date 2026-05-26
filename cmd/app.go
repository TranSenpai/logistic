package main

import (
	middlewares "goBackend/internal/common/middleware"
	dependency "goBackend/internal/di"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

type App struct {
	engine *gin.Engine
}

func NewApp() (*App, error) {
	// Multiplexer / Dispatcher (Bộ phân luồng và định tuyến)
	// Dùng Default trả về 1 Dispatcher có Logger và Recovery(nếu server bị sập thì retry) trước
	ginEngine := gin.Default()
	ginEngine.Use(middlewares.ErrorHandler())

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://localhost:3000", "http://localhost:8080"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	corsConfig.ExposeHeaders = []string{"Content-Length"}
	corsConfig.AllowCredentials = true
	corsConfig.MaxAge = 12 * time.Hour
	ginEngine.Use(cors.New(corsConfig))

	err := dependency.Injection(ginEngine)
	if err != nil {
		return nil, err
	}

	return &App{engine: ginEngine}, nil
}

func (a *App) Start() error {
	return a.engine.Run()
}
