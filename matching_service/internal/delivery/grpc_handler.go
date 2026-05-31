package delivery

import (
	logisticsv1 "goBackend/api/logistics/v1/gen/go/logistics/v1"
	"matching_service/internal/biz"
)

// LogisticsDelivery implements the gRPC interface defined in logistics.proto
// It acts as the boundary layer protecting the Biz logic.
type LogisticsDelivery struct {
	logisticsv1.UnimplementedMatchingEngineServiceServer
	matchingEngine biz.MatchingEngine
}

func NewLogisticsDelivery(engine biz.MatchingEngine) *LogisticsDelivery {
	return &LogisticsDelivery{
		matchingEngine: engine,
	}
}
