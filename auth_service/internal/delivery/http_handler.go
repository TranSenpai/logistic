package delivery

import (
	"auth_service/internal/biz"
	"auth_service/internal/mapper"
	"auth_service/internal/mapper/generated"

	"github.com/gin-gonic/gin"
)

type HttpHandler struct {
	authService biz.AuthService
	authMapper  mapper.AuthMapper
}

func NewHttpHandler(authService biz.AuthService) *HttpHandler {
	return &HttpHandler{
		authService: authService,
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
