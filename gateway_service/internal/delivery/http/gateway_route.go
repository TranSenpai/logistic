package http

import (
	"gateway_service/internal/controller"

	pbauth "github.com/logistic/api/logistic/auth_service/v1"
	pbmatching "github.com/logistic/api/logistic/matching_service/v1"
	pbmedia "github.com/logistic/api/logistic/media_service/v1"
	pbuser "github.com/logistic/api/logistic/user_service/v1"
	pbvehicle "github.com/logistic/api/logistic/vehicle_service/v1"

	_ "gateway_service/docs" // Import swagger docs

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func RegisterGatewayRoutes(
	ginEngine *gin.Engine,
	authClient pbauth.AuthServiceClient,
	matchingClient pbmatching.MatchingEngineServiceClient,
	mediaClient pbmedia.MediaServiceClient,
	userClient pbuser.UserServiceClient,
	vehicleClient pbvehicle.VehicleServiceClient,
) {
	authController := controller.NewAuthController(authClient)
	mediaController := controller.NewMediaController(mediaClient)
	matchingController := controller.NewMatchingController(matchingClient)
	userController := controller.NewUserController(userClient)
	vehicleController := controller.NewVehicleController(vehicleClient)

	// Swagger UI route
	ginEngine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	apiGroup := ginEngine.Group("/api")
	{
		// -----------------------------------------------------------
		// 1. AUTH SERVICE
		// -----------------------------------------------------------
		authGroup := apiGroup.Group("/auth/v1")
		{
			authGroup.POST("/register", authController.Register)
			authGroup.POST("/login", authController.Login)
			authGroup.GET("/get-info", authController.GetInfo)

			// OAuth2 Routes
			authGroup.GET("/google/login", authController.GoogleLogin)
			authGroup.GET("/google/callback", authController.GoogleCallback)
		}

		// -----------------------------------------------------------
		// 2. MEDIA SERVICE
		// -----------------------------------------------------------
		mediaGroup := apiGroup.Group("/media/v1")
		{
			mediaGroup.POST("/upload", mediaController.UploadFile)
			mediaGroup.DELETE("/delete/:publicID", mediaController.DeleteFile)
		}

		// -----------------------------------------------------------
		// 3. MATCHING SERVICE
		// -----------------------------------------------------------
		matchingGroup := apiGroup.Group("/matching/v1")
		{
			matchingGroup.POST("/bid", matchingController.SubmitBid)
			matchingGroup.POST("/ask", matchingController.SubmitAsk)
			matchingGroup.POST("/accept", matchingController.AcceptMatch)
		}

		// -----------------------------------------------------------
		// 4. USER SERVICE
		// -----------------------------------------------------------
		userGroup := apiGroup.Group("/user/v1")
		{
			userGroup.POST("/register", userController.RegisterUser)
			userGroup.GET("/:id", userController.GetUser)
			userGroup.PUT("/:user_id/kyc", userController.UpdateDriverKYC)
		}

		// -----------------------------------------------------------
		// 5. VEHICLE SERVICE
		// -----------------------------------------------------------
		vehicleGroup := apiGroup.Group("/vehicle/v1")
		{
			vehicleGroup.POST("/register", vehicleController.RegisterVehicle)
			vehicleGroup.GET("/list", vehicleController.ListVehicles)
			vehicleGroup.GET("/:id", vehicleController.GetVehicle)
			vehicleGroup.PUT("/:id/status", vehicleController.UpdateVehicleStatus)
		}
	}
}
