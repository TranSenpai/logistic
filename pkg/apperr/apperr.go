// Package apperr là "bảng mã lỗi" dùng chung cho toàn bộ microservice.
//
// Ý tưởng: tầng biz/repo KHÔNG được biết gì về gRPC hay HTTP. Chúng chỉ trả về
// *Error của package này. Việc dịch sang gRPC code (ở interceptor của service)
// hay HTTP status (ở middleware của gateway) là việc của tầng ngoài cùng.
// Nhờ vậy một lỗi "không tìm thấy user" luôn ra 404 ở mọi service mà không phải
// copy-paste if/else ở từng controller.
package apperr

import (
	"errors"
	"fmt"
	"net/http"

	"google.golang.org/grpc/codes"
)

// Kind phân loại lỗi theo NGỮ NGHĨA nghiệp vụ, không theo transport.
type Kind string

const (
	KindInvalidArgument  Kind = "INVALID_ARGUMENT"
	KindUnauthenticated  Kind = "UNAUTHENTICATED"
	KindPermissionDenied Kind = "PERMISSION_DENIED"
	KindNotFound         Kind = "NOT_FOUND"
	KindAlreadyExists    Kind = "ALREADY_EXISTS"
	KindConflict         Kind = "CONFLICT"
	KindFailedPrecond    Kind = "FAILED_PRECONDITION"
	KindResourceExceeded Kind = "RESOURCE_EXHAUSTED"
	KindUnavailable      Kind = "UNAVAILABLE"
	KindTimeout          Kind = "DEADLINE_EXCEEDED"
	KindInternal         Kind = "INTERNAL"
)

// Error là kiểu lỗi chuẩn của hệ thống.
//
//	Kind    -> quyết định gRPC code / HTTP status
//	Code    -> mã máy đọc được, client dùng để switch (vd: "USER_NOT_FOUND")
//	Message -> câu chữ trả cho client (đã an toàn để lộ ra ngoài)
//	Details -> metadata phụ (field nào sai, id nào...)
//	cause   -> lỗi gốc, GIỮ LẠI để log nhưng KHÔNG trả ra ngoài
type Error struct {
	Kind    Kind
	Code    string
	Message string
	Details map[string]string
	cause   error
}

func (e *Error) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s(%s): %s: %v", e.Kind, e.Code, e.Message, e.cause)
	}
	return fmt.Sprintf("%s(%s): %s", e.Kind, e.Code, e.Message)
}

// Unwrap cho phép errors.Is/errors.As xuyên qua lớp bọc.
func (e *Error) Unwrap() error { return e.cause }

// WithCause gắn lỗi gốc (sql, redis, grpc...) để interceptor log lại.
func (e *Error) WithCause(err error) *Error {
	clone := *e
	clone.cause = err
	return &clone
}

// WithDetail thêm một cặp metadata. An toàn khi Details đang nil.
func (e *Error) WithDetail(key, value string) *Error {
	clone := *e
	clone.Details = make(map[string]string, len(e.Details)+1)
	for k, v := range e.Details {
		clone.Details[k] = v
	}
	clone.Details[key] = value
	return &clone
}

// WithMessage thay câu chữ nhưng giữ nguyên Kind/Code.
func (e *Error) WithMessage(format string, args ...any) *Error {
	clone := *e
	clone.Message = fmt.Sprintf(format, args...)
	return &clone
}

// New tạo một mã lỗi mới. Thường được gọi ở biến package-level của từng service.
func New(kind Kind, code, message string) *Error {
	return &Error{Kind: kind, Code: code, Message: message}
}

func InvalidArgument(code, msg string) *Error    { return New(KindInvalidArgument, code, msg) }
func NotFound(code, msg string) *Error           { return New(KindNotFound, code, msg) }
func AlreadyExists(code, msg string) *Error      { return New(KindAlreadyExists, code, msg) }
func Conflict(code, msg string) *Error           { return New(KindConflict, code, msg) }
func FailedPrecondition(code, msg string) *Error { return New(KindFailedPrecond, code, msg) }
func PermissionDenied(code, msg string) *Error   { return New(KindPermissionDenied, code, msg) }
func Unauthenticated(code, msg string) *Error    { return New(KindUnauthenticated, code, msg) }
func Unavailable(code, msg string) *Error        { return New(KindUnavailable, code, msg) }
func Internal(code, msg string) *Error           { return New(KindInternal, code, msg) }

// From bóc *Error ra khỏi chuỗi wrap. Nếu err không phải lỗi của hệ thống,
// trả về false để tầng ngoài biết đây là lỗi "lạ" (cần log full stack, trả 500).
func From(err error) (*Error, bool) {
	var e *Error
	if errors.As(err, &e) {
		return e, true
	}
	return nil, false
}

// GRPCCode dịch Kind sang gRPC status code.
func (e *Error) GRPCCode() codes.Code {
	switch e.Kind {
	case KindInvalidArgument:
		return codes.InvalidArgument
	case KindUnauthenticated:
		return codes.Unauthenticated
	case KindPermissionDenied:
		return codes.PermissionDenied
	case KindNotFound:
		return codes.NotFound
	case KindAlreadyExists:
		return codes.AlreadyExists
	case KindConflict:
		return codes.Aborted
	case KindFailedPrecond:
		return codes.FailedPrecondition
	case KindResourceExceeded:
		return codes.ResourceExhausted
	case KindUnavailable:
		return codes.Unavailable
	case KindTimeout:
		return codes.DeadlineExceeded
	default:
		return codes.Internal
	}
}

// HTTPStatus dịch Kind sang HTTP status, dùng ở gateway.
func (e *Error) HTTPStatus() int {
	switch e.Kind {
	case KindInvalidArgument:
		return http.StatusBadRequest
	case KindUnauthenticated:
		return http.StatusUnauthorized
	case KindPermissionDenied:
		return http.StatusForbidden
	case KindNotFound:
		return http.StatusNotFound
	case KindAlreadyExists, KindConflict:
		return http.StatusConflict
	case KindFailedPrecond:
		return http.StatusUnprocessableEntity
	case KindResourceExceeded:
		return http.StatusTooManyRequests
	case KindUnavailable:
		return http.StatusServiceUnavailable
	case KindTimeout:
		return http.StatusGatewayTimeout
	default:
		return http.StatusInternalServerError
	}
}

// HTTPStatusFromGRPC dùng ở gateway khi chỉ còn cầm gRPC code (đã qua dây).
func HTTPStatusFromGRPC(c codes.Code) int {
	switch c {
	case codes.OK:
		return http.StatusOK
	case codes.InvalidArgument, codes.OutOfRange:
		return http.StatusBadRequest
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists, codes.Aborted:
		return http.StatusConflict
	case codes.FailedPrecondition:
		return http.StatusUnprocessableEntity
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	case codes.Unimplemented:
		return http.StatusNotImplemented
	default:
		return http.StatusInternalServerError
	}
}
