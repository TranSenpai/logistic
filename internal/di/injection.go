package di

import (
	"os"

	"goBackend/internal/biz"
	entclient "goBackend/internal/common/ent_client"
	"goBackend/internal/controller"
	"goBackend/internal/delivery"
	"goBackend/internal/repo"

	"github.com/gin-gonic/gin"
)

func Injection(ginEngine *gin.Engine) error {
	clientDb, err := entclient.NewConnection()
	if err != nil {
		return err
	}

	articleRepo := repo.NewArticleRepo(clientDb)
	articleUsecase := biz.NewArticleUsecase(articleRepo)
	articleController := controller.NewArticleController(articleUsecase)
	articleDelivery := delivery.NewDelivery(articleController)
	routeDelivery := delivery.NewRouteDelivery(articleDelivery)
	routeDelivery.RegisterRouter(ginEngine)

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "dev_secret_change_before_production"
	}

	authRepo := repo.NewAuthRepo(clientDb)
	authService := biz.NewAuthService(authRepo, jwtSecret)
	authController := controller.NewAuthController(authService)
	authDelivery := delivery.NewAuthDelivery(authController)

	apiGroup := ginEngine.Group("/api/v1")
	authDelivery.RegisterRouter(apiGroup)

	return nil
}
