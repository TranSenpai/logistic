package http

import (
	"gateway_service/internal/handler"

	pb "github.com/logistic/api/logistic/auth_service/v1"

	"github.com/gin-gonic/gin"
)

// RegisterGatewayRoutes đăng ký tất cả các route của Gateway Service
func RegisterGatewayRoutes(ginEngine *gin.Engine, authClient pb.AuthServiceClient) {
	// Khởi tạo các handler
	authHandler := handler.NewAuthHandler(authClient)

	// API Group cha
	apiGroup := ginEngine.Group("/api/v1")
	{
		// Group Auth
		authGroup := apiGroup.Group("/auth")
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)
			authGroup.GET("/get-info", authHandler.GetInfo)

			// OAuth2 Routes
			authGroup.GET("/google/login", authHandler.GoogleLogin)
			authGroup.GET("/google/callback", authHandler.GoogleCallback)
		}

		// Sau này có thể thêm các service khác ở đây, ví dụ:
		// vehicleGroup := apiGroup.Group("/vehicle")
		// ...
	}
}
