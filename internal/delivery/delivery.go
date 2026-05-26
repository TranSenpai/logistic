package delivery

import (
	controller "goBackend/internal/controller"

	"github.com/gin-gonic/gin"
)

type articleDelivery struct {
	articleController *controller.ArticleController
}

func NewDelivery(articleController *controller.ArticleController) *articleDelivery {
	return &articleDelivery{
		articleController: articleController,
	}
}

func (d *articleDelivery) RegisterRouterV1(apiGroup *gin.RouterGroup) {
	articleGroup := apiGroup.Group("v1/article")
	{
		articleGroup.POST("", d.articleController.Create)
		articleGroup.GET("", d.articleController.Get)
		articleGroup.PUT("/:id", d.articleController.Update)
		articleGroup.DELETE("/:id", d.articleController.Delete)

	}
}
