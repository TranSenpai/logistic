package middleware

import (
	"context"
	"errors"
	"wallet_service/ent"
	"wallet_service/internal/biz"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrorHandlerInterceptor intercepts all gRPC calls and maps internal errors to gRPC status codes.
func ErrorHandlerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		res, err := handler(ctx, req)
		if err != nil {
			return res, mapErrorToGrpcStatus(err)
		}
		return res, nil
	}
}

func mapErrorToGrpcStatus(err error) error {
	// Lỗi nghiệp vụ
	if errors.Is(err, biz.ErrInvalidAmount) || errors.Is(err, biz.ErrSelfTransfer) {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	if errors.Is(err, biz.ErrInsufficientBalance) {
		return status.Error(codes.FailedPrecondition, err.Error())
	}

	// Lỗi Data
	if errors.Is(err, biz.ErrWalletNotFound) || ent.IsNotFound(err) || errors.Is(err, biz.ErrSystemWalletNotFound) {
		return status.Error(codes.NotFound, err.Error())
	}

	// Lỗi mặc định (Hệ thống)
	return status.Errorf(codes.Internal, "internal server error: %v", err)
}
