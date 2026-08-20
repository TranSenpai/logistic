// Package http là BẢN ĐỒ đường đi của toàn hệ thống.
//
// Kiến trúc: Nginx -> gateway_service (file này) -> gRPC tới internal service.
// Client bên ngoài KHÔNG bao giờ chạm trực tiếp vào service nội bộ.
//
// Cây route được chia làm hai nhánh, và đó là điểm thiết kế quan trọng nhất ở đây:
//
//	/api/v1/...        -> app tài xế và app chủ hàng.
//	/api/v1/admin/...  -> trang quản trị, gắn RequireRole("admin") ở CẤP GROUP.
//
// Gắn quyền ở cấp group thay vì từng handler nghĩa là không thể "quên" bảo vệ
// một endpoint admin mới: chỉ cần khai báo nó trong adminGroup là đã được chặn.
package http

import (
	"net/http"

	"gateway_service/internal/controller"
	"gateway_service/internal/middleware"

	pbauth "github.com/logistic/api/logistic/auth_service/v1"
	pbmatching "github.com/logistic/api/logistic/matching_service/v1"
	pbmedia "github.com/logistic/api/logistic/media_service/v1"
	pbnotification "github.com/logistic/api/logistic/notification_service/v1"
	pbuser "github.com/logistic/api/logistic/user_service/v1"
	pbvehicle "github.com/logistic/api/logistic/vehicle_service/v1"

	_ "gateway_service/docs" // swagger docs

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Clients gom mọi gRPC client lại để chữ ký hàm không phình ra theo số service.
type Clients struct {
	Auth         pbauth.AuthServiceClient
	Matching     pbmatching.MatchingEngineServiceClient
	Media        pbmedia.MediaServiceClient
	User         pbuser.UserServiceClient
	Vehicle      pbvehicle.VehicleServiceClient
	Notification pbnotification.NotificationServiceClient
}

func RegisterGatewayRoutes(engine *gin.Engine, clients Clients) {
	// Thứ tự middleware có ý nghĩa — xem ghi chú trong package middleware.
	engine.Use(
		middleware.RequestID(),
		middleware.Recovery(),
		middleware.AccessLog(),
		middleware.IdentityContext(),
		middleware.ErrorGuard(),
	)

	authController := controller.NewAuthController(clients.Auth)
	mediaController := controller.NewMediaController(clients.Media)
	matchingController := controller.NewMatchingController(clients.Matching)
	userController := controller.NewUserController(clients.User)
	vehicleController := controller.NewVehicleController(clients.Vehicle)
	notifController := controller.NewNotificationController(clients.Notification)

	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check cho load balancer / docker healthcheck.
	engine.GET("/healthz", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := engine.Group("/api/v1")

	// =======================================================================
	// AUTH  (5 endpoint)
	// =======================================================================
	auth := api.Group("/auth")
	{
		auth.POST("/register", authController.Register)
		auth.POST("/login", authController.Login)
		auth.GET("/me", authController.GetInfo)
		auth.GET("/google/login", authController.GoogleLogin)
		auth.GET("/google/callback", authController.GoogleCallback)
	}

	// =======================================================================
	// MEDIA  (2 endpoint)
	// =======================================================================
	media := api.Group("/media")
	{
		media.POST("/upload", mediaController.UploadFile)
		media.DELETE("/files/:publicID", mediaController.DeleteFile)
	}

	// =======================================================================
	// USER  (8 endpoint)
	// =======================================================================
	users := api.Group("/users")
	{
		users.POST("/register", userController.RegisterUser)
		users.GET("/:user_id", userController.GetUser)
		users.PUT("/:user_id", userController.UpdateUser)

		users.GET("/:user_id/driver-profile", userController.GetDriverProfile)
		users.PUT("/:user_id/driver-profile", userController.UpdateDriverProfile)
		users.GET("/:user_id/shipper-profile", userController.GetShipperProfile)
		users.PUT("/:user_id/shipper-profile", userController.UpdateShipperProfile)
		users.PUT("/:user_id/kyc", userController.UpdateDriverKYC)

		// --- Sổ địa chỉ (2 endpoint nằm trong nhánh user) ---
		users.POST("/:user_id/addresses", userController.CreateAddress)
		users.GET("/:user_id/addresses", userController.ListAddresses)

		// --- Thiết bị nhận push (2 endpoint) ---
		users.POST("/:user_id/devices", userController.RegisterDevice)
		users.GET("/:user_id/devices", userController.ListDevices)

		// --- Hộp thư thông báo (5 endpoint) ---
		users.GET("/:user_id/notifications", notifController.ListNotifications)
		users.PUT("/:user_id/notifications/read-all", notifController.MarkAllAsRead)
		users.GET("/:user_id/notifications/unread-count", notifController.GetUnreadCount)
		users.GET("/:user_id/notification-preferences", notifController.GetPreferences)
		users.PUT("/:user_id/notification-preferences", notifController.UpdatePreferences)
	}

	// Địa chỉ và thiết bị thao tác theo id riêng của chúng, không lồng dưới user.
	addresses := api.Group("/addresses")
	{
		addresses.PUT("/:id", userController.UpdateAddress)
		addresses.DELETE("/:id", userController.DeleteAddress)
	}

	devices := api.Group("/devices")
	{
		devices.DELETE("/:id", userController.DeleteDevice)
	}

	notifications := api.Group("/notifications")
	{
		notifications.GET("/:id", notifController.GetNotification)
		notifications.PUT("/:id/read", notifController.MarkAsRead)
		notifications.DELETE("/:id", notifController.DeleteNotification)
	}

	// =======================================================================
	// VEHICLE  (12 endpoint)
	// =======================================================================
	vehicles := api.Group("/vehicles")
	{
		vehicles.POST("", vehicleController.RegisterVehicle)
		vehicles.GET("", vehicleController.ListVehicles)

		// Đăng ký TRƯỚC route có tham số: gin ưu tiên đoạn tĩnh, nên "nearby"
		// không bị nuốt bởi "/:id". Đảo thứ tự là gin báo xung đột wildcard.
		vehicles.POST("/nearby", vehicleController.SearchNearbyVehicles)

		vehicles.GET("/:id", vehicleController.GetVehicle)
		vehicles.PUT("/:id", vehicleController.UpdateVehicle)
		vehicles.DELETE("/:id", vehicleController.DeleteVehicle)
		vehicles.PUT("/:id/status", vehicleController.UpdateVehicleStatus)

		vehicles.POST("/:id/documents", vehicleController.UploadVehicleDocument)
		vehicles.GET("/:id/documents", vehicleController.ListVehicleDocuments)

		vehicles.POST("/:id/location", vehicleController.ReportLocation)
		vehicles.GET("/:id/location", vehicleController.GetVehicleLocation)
	}

	vehicleDocs := api.Group("/vehicle-documents")
	{
		vehicleDocs.DELETE("/:id", vehicleController.DeleteVehicleDocument)
	}

	drivers := api.Group("/drivers")
	{
		drivers.POST("/:driver_id/availability", vehicleController.SetDriverAvailability)
		drivers.GET("/:driver_id/availability", vehicleController.GetDriverAvailability)
	}

	// =======================================================================
	// MATCHING  (5 endpoint)
	// =======================================================================
	matching := api.Group("/matching")
	{
		matching.POST("/bids", matchingController.SubmitBid)
		matching.POST("/asks", matchingController.SubmitAsk)
		matching.POST("/offers", matchingController.SubmitOffer)
		matching.POST("/offers/reject", matchingController.RejectOffer)
		matching.POST("/matches/accept", matchingController.AcceptMatch)
	}

	// =======================================================================
	// ADMIN  (16 endpoint) — toàn bộ nhóm này yêu cầu vai trò admin
	// =======================================================================
	admin := api.Group("/admin", middleware.RequireRole("admin"))
	{
		adminUsers := admin.Group("/users")
		{
			adminUsers.GET("", userController.AdminListUsers)
			adminUsers.GET("/stats", userController.AdminGetUserStats)
			adminUsers.PUT("/:id/status", userController.AdminUpdateUserStatus)
			adminUsers.DELETE("/:id", userController.AdminDeleteUser)
		}

		adminKyc := admin.Group("/kyc")
		{
			adminKyc.GET("/pending", userController.AdminListPendingKYC)
			adminKyc.PUT("/:user_id/review", userController.AdminReviewKYC)
		}

		adminVehicles := admin.Group("/vehicles")
		{
			adminVehicles.GET("", vehicleController.AdminListVehicles)
			adminVehicles.GET("/stats", vehicleController.AdminGetVehicleStats)
			adminVehicles.PUT("/:id/verify", vehicleController.AdminVerifyVehicle)
		}

		adminVehicleDocs := admin.Group("/vehicle-documents")
		{
			adminVehicleDocs.GET("/pending", vehicleController.AdminListPendingDocuments)
			adminVehicleDocs.PUT("/:id/review", vehicleController.AdminReviewDocument)
		}

		adminNotifications := admin.Group("/notifications")
		{
			adminNotifications.GET("", notifController.AdminListNotifications)
			adminNotifications.GET("/stats", notifController.AdminGetNotificationStats)
			adminNotifications.POST("/send", notifController.AdminSendNotification)
		}

		adminTemplates := admin.Group("/notification-templates")
		{
			adminTemplates.GET("", notifController.AdminListTemplates)
			adminTemplates.POST("", notifController.AdminCreateTemplate)
			adminTemplates.PUT("/:id", notifController.AdminUpdateTemplate)
			adminTemplates.DELETE("/:id", notifController.AdminDeleteTemplate)
		}
	}
}
