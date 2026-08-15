package di

import (
	"context"
	"log"
	"os"

	"gateway_service/internal/delivery/http"

	pbauth "github.com/logistic/api/logistic/auth_service/v1"
	pbmatching "github.com/logistic/api/logistic/matching_service/v1"
	pbuser "github.com/logistic/api/logistic/user_service/v1"
	pbvehicle "github.com/logistic/api/logistic/vehicle_service/v1"

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
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

	matchingGrpcAddr := os.Getenv("GATEWAY_MATCHING_GRPC_ADDR")
	if matchingGrpcAddr == "" {
		matchingGrpcAddr = "matching_service:9002"
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
		userGrpcAddr = "user_service:9003"
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
	// Vehicle Service gRPC Client
	vehicleGrpcAddr := os.Getenv("GATEWAY_VEHICLE_GRPC_ADDR")
	if vehicleGrpcAddr == "" {
		vehicleGrpcAddr = "vehicle_service:9004"
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

	// Register các HTTP route (cũ)
	http.RegisterGatewayRoutes(ginEngine, authClient, matchingClient)

	// Khởi tạo gRPC-Gateway Mux cho các service mới
	gwMux := runtime.NewServeMux()
	
	// Register gRPC-Gateway handlers
	err = pbuser.RegisterUserServiceHandler(context.Background(), gwMux, userConn)
	if err != nil {
		log.Printf("Failed to register user_service gateway: %v", err)
		return err
	}

	err = pbvehicle.RegisterVehicleServiceHandler(context.Background(), gwMux, vehicleConn)
	if err != nil {
		log.Printf("Failed to register vehicle_service gateway: %v", err)
		return err
	}

	// Tích hợp gRPC-Gateway vào Gin
	ginEngine.Any("/v1/*any", gin.WrapH(gwMux))

	return nil
}
