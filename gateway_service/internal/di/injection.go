package di

import (
	"log"
	"os"

	"gateway_service/internal/delivery/http"

	pbauth "github.com/logistic/api/logistic/auth_service/v1"
	pbmatching "github.com/logistic/api/logistic/matching_service/v1"
	pbmedia "github.com/logistic/api/logistic/media_service/v1"
	pbuser "github.com/logistic/api/logistic/user_service/v1"
	pbvehicle "github.com/logistic/api/logistic/vehicle_service/v1"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
)

func Injection(ginEngine *gin.Engine) error {
	authGrpcAddr := os.Getenv("GATEWAY_AUTH_GRPC_ADDR")
	if authGrpcAddr == "" {
		authGrpcAddr = "localhost:9001" // Fallback cho local dev
	}

	conn, err := grpc.NewClient(
		authGrpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		log.Printf("Failed to connect to auth_service via gRPC: %v", err)
		return err
	}
	// Note: We don't defer conn.Close() here because it needs to stay open for the lifetime of the application.
	// You might want to handle graceful shutdown at the App level.

	authClient := pbauth.NewAuthServiceClient(conn)

	mediaGrpcAddr := os.Getenv("GATEWAY_MEDIA_GRPC_ADDR")
	if mediaGrpcAddr == "" {
		mediaGrpcAddr = "media-service:9002"
	}
	mediaConn, err := grpc.NewClient(
		mediaGrpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		log.Printf("Failed to connect to media_service via gRPC: %v", err)
		return err
	}
	mediaClient := pbmedia.NewMediaServiceClient(mediaConn)

	matchingGrpcAddr := os.Getenv("GATEWAY_MATCHING_GRPC_ADDR")
	if matchingGrpcAddr == "" {
		matchingGrpcAddr = "matching-service:9003"
	}

	matchingConn, err := grpc.NewClient(
		matchingGrpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		log.Printf("Failed to connect to matching_service via gRPC: %v", err)
		return err
	}
	matchingClient := pbmatching.NewMatchingEngineServiceClient(matchingConn)

	// User Service gRPC Client
	userGrpcAddr := os.Getenv("GATEWAY_USER_GRPC_ADDR")
	if userGrpcAddr == "" {
		userGrpcAddr = "user-service:9004"
	}
	userConn, err := grpc.NewClient(
		userGrpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		log.Printf("Failed to connect to user_service via gRPC: %v", err)
		return err
	}
	userClient := pbuser.NewUserServiceClient(userConn)

	// Vehicle Service gRPC Client
	vehicleGrpcAddr := os.Getenv("GATEWAY_VEHICLE_GRPC_ADDR")
	if vehicleGrpcAddr == "" {
		vehicleGrpcAddr = "vehicle-service:9005"
	}
	vehicleConn, err := grpc.NewClient(
		vehicleGrpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		log.Printf("Failed to connect to vehicle_service via gRPC: %v", err)
		return err
	}
	vehicleClient := pbvehicle.NewVehicleServiceClient(vehicleConn)

	// Register các HTTP route (cũ)
	http.RegisterGatewayRoutes(ginEngine, authClient, matchingClient, mediaClient, userClient, vehicleClient)

	return nil
}
