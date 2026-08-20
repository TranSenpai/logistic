// Package middleware chứa các interceptor/middleware dùng chung cho mọi service.
//
// Vì sao phải có: trước đây mỗi controller tự viết `if err != nil { return nil, err }`,
// nên lỗi *ent.NotFoundError rò thẳng ra client dưới dạng codes.Unknown kèm câu
// tiếng Anh của thư viện. Interceptor ở đây là CHỐT CHẶN CUỐI: mọi lỗi đi ra khỏi
// service đều được chuẩn hoá thành gRPC status + ErrorInfo có mã máy đọc được,
// còn chi tiết kỹ thuật thì chỉ nằm lại trong log.
package middleware

import (
	"context"
	"log"
	"runtime/debug"
	"time"

	"github.com/logistic/pkg/apperr"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrorInterceptor dịch mọi lỗi trả về từ handler thành gRPC status chuẩn.
//
// Thứ tự ưu tiên:
//  1. Đã là *apperr.Error  -> dùng Kind/Code của nó.
//  2. Đã là gRPC status    -> giữ nguyên (thường là lỗi vọng lên từ service khác).
//  3. Còn lại              -> codes.Internal + câu chữ chung chung, log full lỗi gốc.
func ErrorInterceptor(serviceName string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		resp, err := handler(ctx, req)
		if err == nil {
			return resp, nil
		}
		return nil, toStatusError(serviceName, info.FullMethod, err)
	}
}

func toStatusError(serviceName, method string, err error) error {
	if appErr, ok := apperr.From(err); ok {
		st := status.New(appErr.GRPCCode(), appErr.Message)

		info := &errdetails.ErrorInfo{
			Reason:   appErr.Code,
			Domain:   serviceName,
			Metadata: appErr.Details,
		}
		if withDetails, dErr := st.WithDetails(info); dErr == nil {
			st = withDetails
		}

		// Lỗi nghiệp vụ (404, 400...) là chuyện bình thường -> log mức nhẹ.
		// Chỉ lỗi Internal mới đáng báo động.
		if appErr.Kind == apperr.KindInternal {
			log.Printf("[%s][ERROR] %s -> %v", serviceName, method, appErr.Error())
		}
		return st.Err()
	}

	// Lỗi đã mang sẵn gRPC status: giữ nguyên code để không "nuốt" mất ngữ nghĩa.
	if st, ok := status.FromError(err); ok && st.Code() != codes.Unknown {
		return err
	}

	// Lỗi lạ: KHÔNG trả nguyên văn ra ngoài (tránh lộ tên bảng, câu SQL...).
	log.Printf("[%s][UNHANDLED] %s -> %v", serviceName, method, err)
	return status.Error(codes.Internal, "internal server error")
}

// RecoveryInterceptor biến panic thành lỗi Internal thay vì làm sập cả process.
// Một request lỗi không được phép kéo theo toàn bộ service.
func RecoveryInterceptor(serviceName string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[%s][PANIC] %s -> %v\n%s", serviceName, info.FullMethod, r, debug.Stack())
				err = status.Error(codes.Internal, "internal server error")
				resp = nil
			}
		}()
		return handler(ctx, req)
	}
}

// LoggingInterceptor ghi lại method + thời gian + code trả về.
func LoggingInterceptor(serviceName string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		log.Printf("[%s] %s code=%s dur=%s", serviceName, info.FullMethod, status.Code(err), time.Since(start))
		return resp, err
	}
}

// ChainForService trả về bộ interceptor chuẩn theo đúng thứ tự nên dùng.
//
// Thứ tự CÓ Ý NGHĨA: Recovery phải nằm ngoài cùng để bắt được cả panic xảy ra
// bên trong Error/Logging; Error nằm sát handler nhất để nó là thứ cuối cùng
// chạm vào error trước khi error rời khỏi service.
func ChainForService(serviceName string) grpc.ServerOption {
	return grpc.ChainUnaryInterceptor(
		RecoveryInterceptor(serviceName),
		LoggingInterceptor(serviceName),
		ErrorInterceptor(serviceName),
	)
}
