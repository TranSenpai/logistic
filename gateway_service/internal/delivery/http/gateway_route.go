package http

import (
	"gateway_service/internal/handler"

	pbauth "github.com/logistic/api/logistic/auth_service/v1"
	pbmatching "github.com/logistic/api/logistic/matching_service/v1"
	pbmedia "github.com/logistic/api/logistic/media_service/v1"

	_ "gateway_service/docs" // Import swagger docs

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func RegisterGatewayRoutes(ginEngine *gin.Engine, authClient pbauth.AuthServiceClient, matchingClient pbmatching.MatchingEngineServiceClient, mediaClient pbmedia.MediaServiceClient) {
	authHandler := handler.NewAuthHandler(authClient)
	// matchingHandler := handler.NewMatchingHandler(matchingClient) // TODO: Implement matching handler

	// Swagger UI route
	ginEngine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

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

		mediaHandler := handler.NewMediaHandler(mediaClient)
		mediaGroup := apiGroup.Group("/media")
		{
			mediaGroup.POST("/upload", mediaHandler.UploadFile)
			mediaGroup.DELETE("/delete/:publicID", mediaHandler.DeleteFile)
		}
	}
}
