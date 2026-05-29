package di

import (
	entclient "goBackend/matching_service/internal/common/ent_client"

	"github.com/gin-gonic/gin"
)

func Injection(ginEngine *gin.Engine) error {
	_, err := entclient.NewConnection()
	if err != nil {
		return err
	}

	// Matching Service only creates the engine and delivery.
	// matchingEngine := biz.NewMatchingEngine()
	// delivery.NewLogisticsDelivery(matchingEngine) // GRPC delivery would be registered here

	return nil
}
