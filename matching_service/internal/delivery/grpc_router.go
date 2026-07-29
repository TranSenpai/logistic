package delivery

import (
	logisticsv1 "goBackend/api/logistic/matching/v1"
	"matching_service/internal/handler/grpchandler"

	"google.golang.org/grpc"
)

// RegisterGrpcRouter is responsible ONLY for mapping the gRPC Server to our Handler.
func RegisterGrpcRouter(grpcServer *grpc.Server, handler *grpchandler.MatchingHandler) {
	logisticsv1.RegisterMatchingEngineServiceServer(grpcServer, handler)
}
