package delivery

import (
	"goBackend/auth_service/internal/biz"
	"goBackend/auth_service/internal/mapper"
	"goBackend/auth_service/internal/mapper/generated"

	"github.com/gin-gonic/gin"
)

type HttpHandler struct {
	authUsecase biz.AuthUsecase
	authMapper  mapper.AuthMapper
}

func NewHttpHandler(authUsecase biz.AuthUsecase) *HttpHandler {
	return &HttpHandler{
		authUsecase: authUsecase,
		authMapper:  &generated.AuthMapperImpl{},
	}
}

func (h *HttpHandler) RegisterRouter(ginEngine *gin.Engine) {
	apiGroup := ginEngine.Group("/api/v1")
	authGroup := apiGroup.Group("auth")
	{
		authGroup.POST("/register", h.Register)
		authGroup.POST("/login", h.Login)
		authGroup.GET("/get-info", h.GetInfo)

		// OAuth2 Routes
		authGroup.GET("/google/login", h.GoogleLogin)
		authGroup.GET("/google/callback", h.GoogleCallback)
	}
}
