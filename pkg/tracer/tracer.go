package tracer

import (
	"context"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// InitTracer initializes an OTLP exporter, and configures the corresponding trace and
// metric providers.
func InitTracer(serviceName string, collectorURL string) (func(context.Context) error, error) {
	ctx := context.Background()

	// 1. Tạo OTLP gRPC Exporter
	// Nếu không truyền URL, mặc định SDK sẽ gửi tới localhost:4317
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
	}
	if collectorURL != "" {
		opts = append(opts, otlptracegrpc.WithEndpoint(collectorURL))
	}

	exporter, err := otlptracegrpc.New(ctx, opts...)
	if err != nil {
		log.Printf("Failed to create OTLP trace exporter: %v", err)
		return nil, err
	}

	// 2. Định nghĩa Resource (để định danh service trên Jaeger/Otel Collector)
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		log.Printf("Failed to create resource: %v", err)
		return nil, err
	}

	// 3. Khởi tạo Tracer Provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	// 4. Set global Tracer Provider và TextMapPropagator
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Trả về hàm shutdown để gọi lúc tắt app
	return func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		return tp.Shutdown(ctx)
	}, nil
}
