// Package di dựng các kết nối gRPC xuống service nội bộ rồi giao cho tầng route.
//
// Gateway KHÔNG có database và KHÔNG có luật nghiệp vụ: nó chỉ dịch HTTP <-> gRPC,
// chuẩn hoá lỗi và chặn quyền. Mọi thứ khác nằm ở service phía sau.
package di

import (
	"fmt"
	"log"
	"os"

	"gateway_service/internal/delivery/http"

	pbauth "github.com/logistic/api/logistic/auth_service/v1"
	pbmatching "github.com/logistic/api/logistic/matching_service/v1"
	pbmedia "github.com/logistic/api/logistic/media_service/v1"
	pbnotification "github.com/logistic/api/logistic/notification_service/v1"
	pbuser "github.com/logistic/api/logistic/user_service/v1"
	pbvehicle "github.com/logistic/api/logistic/vehicle_service/v1"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Container giữ các kết nối gRPC để đóng gọn lúc tắt gateway.
type Container struct {
	conns []*grpc.ClientConn
}

func (c *Container) Close() {
	if c == nil {
		return
	}
	for _, conn := range c.conns {
		if conn == nil {
			continue
		}
		if err := conn.Close(); err != nil {
			log.Printf("[gateway] closing grpc conn failed: %v", err)
		}
	}
}

// dial mở kết nối gRPC tới một service nội bộ.
//
// grpc.NewClient là lazy: nó KHÔNG bắt tay ngay, mà kết nối ở lần gọi đầu tiên.
// Nhờ vậy gateway khởi động được kể cả khi một service phía sau chưa sẵn sàng —
// điều bình thường khi cả cụm cùng bật lên trong docker-compose.
func dial(name, envKey, fallback string) (*grpc.ClientConn, error) {
	addr := os.Getenv(envKey)
	if addr == "" {
		addr = fallback
	}

	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		return nil, fmt.Errorf("gateway: không tạo được client gRPC tới %s (%s): %w", name, addr, err)
	}

	log.Printf("[gateway] client gRPC %s -> %s", name, addr)
	return conn, nil
}

func Injection(ginEngine *gin.Engine) (*Container, error) {
	container := &Container{}

	type target struct {
		name     string
		envKey   string
		fallback string
	}

	targets := []target{
		{"auth_service", "GATEWAY_AUTH_GRPC_ADDR", "auth-service:9001"},
		{"media_service", "GATEWAY_MEDIA_GRPC_ADDR", "media-service:9002"},
		{"matching_service", "GATEWAY_MATCHING_GRPC_ADDR", "matching-service:9003"},
		{"user_service", "GATEWAY_USER_GRPC_ADDR", "user-service:9004"},
		{"vehicle_service", "GATEWAY_VEHICLE_GRPC_ADDR", "vehicle-service:9005"},
		{"notification_service", "GATEWAY_NOTIFICATION_GRPC_ADDR", "notification-service:9006"},
	}

	conns := make([]*grpc.ClientConn, 0, len(targets))
	for _, t := range targets {
		conn, err := dial(t.name, t.envKey, t.fallback)
		if err != nil {
			container.Close()
			return nil, err
		}
		conns = append(conns, conn)
		container.conns = append(container.conns, conn)
	}

	http.RegisterGatewayRoutes(ginEngine, http.Clients{
		Auth:         pbauth.NewAuthServiceClient(conns[0]),
		Media:        pbmedia.NewMediaServiceClient(conns[1]),
		Matching:     pbmatching.NewMatchingEngineServiceClient(conns[2]),
		User:         pbuser.NewUserServiceClient(conns[3]),
		Vehicle:      pbvehicle.NewVehicleServiceClient(conns[4]),
		Notification: pbnotification.NewNotificationServiceClient(conns[5]),
	})

	return container, nil
}
