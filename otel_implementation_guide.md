# Hướng dẫn triển khai OpenTelemetry (OTel) trong Logistic System

Tài liệu này mô tả chi tiết lộ trình và cách thức tích hợp OpenTelemetry cho distributed tracing trong toàn bộ hệ thống microservices.

## 1. Kiến trúc tổng quan của OpenTelemetry

Trong hệ thống này, chúng ta sử dụng kiến trúc sau cho Observability:
- **Instrumentation**: Thư viện `go.opentelemetry.io/otel` và `otelgrpc` để tự động thu thập trace từ các request gRPC.
- **Exporter**: Gửi data thu thập được qua giao thức **OTLP gRPC** (`OTEL_EXPORTER_OTLP_ENDPOINT`).
- **Collector (Dự kiến)**: Một OpenTelemetry Collector container sẽ nhận trace từ các service và forward sang backend.
- **Backend (Dự kiến)**: Jaeger hoặc Elastic APM để visualize các traces (sẽ được cấu hình trong `docker-compose.yml`).

## 2. Shared Library: `pkg/tracer`

Để tái sử dụng code khởi tạo Otel cho tất cả các services, một module dùng chung đã được tạo tại `pkg/tracer/tracer.go`.

### Thành phần chính trong `tracer.go`:
- Khởi tạo một `otlptracegrpc.Exporter` (OTLP Exporter) chỉ tới endpoint của Collector.
- Cấu hình `resource.Resource` dùng để gán nhãn `service.name` (vd: `user_service`, `gateway_service`) nhằm định danh trace thuộc về service nào.
- Cài đặt Global `TracerProvider` và `TextMapPropagator` (cho phép truyền trace ID context qua các network call giữa client và server thông qua gRPC metadata).

## 3. Cách thức gắn (Instrument) OTel vào một Microservice

### A. Gắn vào gRPC Server (Ví dụ: `user_service`, `vehicle_service`, `matching_service`)

**Bước 1: Khởi tạo Tracer ở tầng `cmd/app.go`**
```go
import "github.com/logistic/pkg/tracer"

func NewApp() *App {
    // 1. Khởi tạo tracer với tên service
    shutdownTracer, err := tracer.InitTracer("user_service", os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
    // ...
```

**Bước 2: Gắn Interceptor vào gRPC Server**
Khi khởi tạo gRPC server, truyền interceptor `otelgrpc.NewServerHandler()` vào thông qua `grpc.StatsHandler`:
```go
import "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

grpcServer := grpc.NewServer(
    grpc.StatsHandler(otelgrpc.NewServerHandler()),
)
```

**Bước 3: Graceful Shutdown Tracer**
Đảm bảo gọi hàm `shutdownTracer` khi service bị stop để đẩy (flush) toàn bộ các trace cuối cùng ra ngoài trước khi process chết:
```go
func (a *App) Stop() {
    // ...
    if a.shutdown != nil {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        a.shutdown(ctx)
    }
}
```

### B. Gắn vào gRPC Client (Ví dụ: `gateway_service`)

Để chuỗi trace không bị đứt đoạn, service gọi gRPC (Client) cũng cần gắn OTel vào `grpc.NewClient`.

**Ví dụ tại `gateway_service/internal/di/injection.go`:**
```go
conn, err := grpc.NewClient(
    authGrpcAddr,
    grpc.WithTransportCredentials(insecure.NewCredentials()),
    grpc.WithStatsHandler(otelgrpc.NewClientHandler()), // Tự động inject TraceID vào gRPC Header
)
```

## 4. Lộ trình triển khai hiện tại và tiếp theo

- [x] Tạo gói thư viện chung `pkg/tracer`.
- [x] Gắn OTel vào gRPC Server: `user_service`, `vehicle_service`, `matching_service`.
- [x] Gắn OTel vào gRPC Client: `gateway_service`.
- [ ] **Sắp tới:** Khi implement `wallet_service` và `notification_service`, cần áp dụng ngay template gắn OTel tương tự như trên.
- [ ] **Sắp tới (Messaging):** Khi tích hợp RabbitMQ/Kafka, cần setup propagate (truyền) trace context vào trong Message Header để có thể trace xuyên suốt từ HTTP -> gRPC -> Message Queue -> Consumer.
- [ ] **Sắp tới (Database):** Cân nhắc gắn plugin Otel cho `ent` framework (nếu cần trace sâu đến từng câu query SQL).
- [ ] **Sắp tới (Infrastructure):** Bổ sung Jaeger/OTel Collector vào file `docker-compose.yml` để xem được giao diện visualizing trace khi chạy trên local.
