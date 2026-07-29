package grpchandler

import (
	logisticsv1 "goBackend/api/logistic/matching/v1"
	"matching_service/internal/biz"
)

// MatchingHandler implements the gRPC interface defined in matching_engine.proto
// It acts as the boundary layer protecting the Biz logic.
type MatchingHandler struct {
	logisticsv1.UnimplementedMatchingEngineServiceServer
	matchingEngine biz.MatchingEngine
}

func NewMatchingHandler(engine biz.MatchingEngine) *MatchingHandler {
	return &MatchingHandler{
		matchingEngine: engine,
	}
}
