package delivery

import (
	"goBackend/internal/controller"

	"github.com/gin-gonic/gin"
)

type authDelivery struct {
	authController *controller.AuthController
}

func NewAuthDelivery(authController *controller.AuthController) *authDelivery {
	return &authDelivery{
		authController: authController,
	}
}

func (d *authDelivery) RegisterRouter(apiGroup *gin.RouterGroup) {
	authGroup := apiGroup.Group("auth")
	{
		authGroup.POST("/register", d.authController.Register)
		authGroup.POST("/login", d.authController.Login)
	}
}
