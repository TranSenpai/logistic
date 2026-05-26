package delivery

import "github.com/gin-gonic/gin"

type RouteDelivery struct {
	articleDelivery *articleDelivery
}

func NewRouteDelivery(articleDelivery *articleDelivery) *RouteDelivery {
	return &RouteDelivery{
		articleDelivery: articleDelivery,
	}
}

func (r *RouteDelivery) RegisterRouter(engine *gin.Engine) {
	apiGroup := engine.Group("api")

	{
		r.articleDelivery.RegisterRouterV1(apiGroup)
	}
}
