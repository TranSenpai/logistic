package main

import (
	"media_service/di"
	"net/http"

	_ "media_service/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type App struct {
	engine *gin.Engine
	server *http.Server
}

func NewApp(port string) (*App, error) {
	r := gin.Default()

	// Đăng ký route cho Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	err := di.Injection(r)
	if err != nil {
		return nil, err
	}

	return &App{
		engine: r,
		server: &http.Server{
			Addr:    ":" + port,
			Handler: r,
		},
	}, nil
}

func (a *App) Start() error {
	return a.server.ListenAndServe()
}
