package di

import (
	"log"
	"os"

	"gateway_service/internal/delivery/http"

	pbauth "github.com/logistic/api/logistic/auth_service/v1"
	pbmatching "github.com/logistic/api/logistic/matching_service/v1"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Injection(ginEngine *gin.Engine) error {
	authGrpcAddr := os.Getenv("AUTH_GRPC_ADDR")
	if authGrpcAddr == "" {
		authGrpcAddr = "auth_service:9001" // fallback mặc định
	}

	conn, err := grpc.NewClient(authGrpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Failed to connect to auth_service via gRPC: %v", err)
		return err
	}
	// Note: We don't defer conn.Close() here because it needs to stay open for the lifetime of the application.
	// You might want to handle graceful shutdown at the App level.

	authClient := pbauth.NewAuthServiceClient(conn)

	matchingGrpcAddr := os.Getenv("MATCHING_GRPC_ADDR")
	if matchingGrpcAddr == "" {
		matchingGrpcAddr = "matching_service:9002"
	}

	matchingConn, err := grpc.NewClient(matchingGrpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Failed to connect to matching_service via gRPC: %v", err)
		return err
	}
	matchingClient := pbmatching.NewMatchingEngineServiceClient(matchingConn)

	// Register các HTTP route
	http.RegisterGatewayRoutes(ginEngine, authClient, matchingClient)

	return nil
}
