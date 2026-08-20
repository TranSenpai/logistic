// Package response chuẩn hoá MỌI phản hồi HTTP đi ra khỏi gateway.
//
// Vấn đề trước đây: mỗi controller tự viết
//
//	st, _ := status.FromError(err)
//	ctx.JSON(http.StatusInternalServerError, gin.H{"error": "...", "message": st.Message()})
//
// Hệ quả là mọi lỗi — kể cả "không tìm thấy user" — đều ra HTTP 500, và mã lỗi
// ("registration_failed", "fetch_failed") do từng người tự nghĩ ra nên client
// không thể xử lý theo mã một cách đáng tin.
//
// Nay: internal service trả gRPC status kèm ErrorInfo (do pkg/middleware gắn),
// gateway đọc lại và dịch sang HTTP status + mã lỗi ỔN ĐỊNH.
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/logistic/pkg/apperr"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HeaderRequestID là tên header mang mã truy vết xuyên suốt một request.
const HeaderRequestID = "X-Request-ID"

// Envelope là khung phản hồi thành công.
type Envelope struct {
	Data      any    `json:"data,omitempty"`
	Message   string `json:"message,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

// ErrorBody là khung phản hồi lỗi.
type ErrorBody struct {
	Error     ErrorDetail `json:"error"`
	RequestID string      `json:"request_id,omitempty"`
}

type ErrorDetail struct {
	// Code là mã máy đọc được, ổn định theo thời gian (USER_NOT_FOUND...).
	// Client nên switch theo trường này, KHÔNG theo Message.
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

func requestID(ctx *gin.Context) string {
	if v, ok := ctx.Get(HeaderRequestID); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ctx.GetHeader(HeaderRequestID)
}

// OK trả 200 kèm dữ liệu.
func OK(ctx *gin.Context, data any) {
	ctx.JSON(http.StatusOK, Envelope{Data: data, RequestID: requestID(ctx)})
}

// OKMessage trả 200 kèm dữ liệu và câu thông báo.
func OKMessage(ctx *gin.Context, data any, message string) {
	ctx.JSON(http.StatusOK, Envelope{Data: data, Message: message, RequestID: requestID(ctx)})
}

// Created trả 201.
func Created(ctx *gin.Context, data any, message string) {
	ctx.JSON(http.StatusCreated, Envelope{Data: data, Message: message, RequestID: requestID(ctx)})
}

// BadRequest dùng cho lỗi phát hiện NGAY TẠI gateway (binding JSON, thiếu tham
// số bắt buộc) — chưa kịp gọi xuống service nào.
func BadRequest(ctx *gin.Context, code, message string) {
	ctx.AbortWithStatusJSON(http.StatusBadRequest, ErrorBody{
		Error:     ErrorDetail{Code: code, Message: message},
		RequestID: requestID(ctx),
	})
}

// Unauthorized / Forbidden dùng cho middleware xác thực.
func Unauthorized(ctx *gin.Context, message string) {
	ctx.AbortWithStatusJSON(http.StatusUnauthorized, ErrorBody{
		Error:     ErrorDetail{Code: "UNAUTHENTICATED", Message: message},
		RequestID: requestID(ctx),
	})
}

func Forbidden(ctx *gin.Context, message string) {
	ctx.AbortWithStatusJSON(http.StatusForbidden, ErrorBody{
		Error:     ErrorDetail{Code: "PERMISSION_DENIED", Message: message},
		RequestID: requestID(ctx),
	})
}

// Error là đường đi CHÍNH của lỗi từ internal service ra client.
//
// Nó bóc ba lớp thông tin theo thứ tự ưu tiên:
//  1. errdetails.ErrorInfo -> mã lỗi nghiệp vụ + metadata (do pkg/middleware gắn).
//  2. gRPC code            -> quyết định HTTP status.
//  3. Message              -> câu chữ hiển thị.
func Error(ctx *gin.Context, err error) {
	st, ok := status.FromError(err)
	if !ok {
		// Không phải lỗi gRPC: nhiều khả năng là lỗi của chính gateway.
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, ErrorBody{
			Error:     ErrorDetail{Code: "INTERNAL", Message: "internal gateway error"},
			RequestID: requestID(ctx),
		})
		return
	}

	detail := ErrorDetail{
		Code:    defaultCodeFor(st.Code()),
		Message: st.Message(),
	}

	for _, d := range st.Details() {
		info, isInfo := d.(*errdetails.ErrorInfo)
		if !isInfo {
			continue
		}
		if info.GetReason() != "" {
			detail.Code = info.GetReason()
		}
		if len(info.GetMetadata()) > 0 {
			detail.Details = info.GetMetadata()
		}
		break
	}

	ctx.AbortWithStatusJSON(apperr.HTTPStatusFromGRPC(st.Code()), ErrorBody{
		Error:     detail,
		RequestID: requestID(ctx),
	})
}

// defaultCodeFor là mã dự phòng khi service không gắn ErrorInfo (ví dụ lỗi phát
// sinh từ chính tầng gRPC: mất kết nối, quá hạn...).
func defaultCodeFor(c codes.Code) string {
	switch c {
	case codes.InvalidArgument:
		return "INVALID_ARGUMENT"
	case codes.NotFound:
		return "NOT_FOUND"
	case codes.AlreadyExists:
		return "ALREADY_EXISTS"
	case codes.PermissionDenied:
		return "PERMISSION_DENIED"
	case codes.Unauthenticated:
		return "UNAUTHENTICATED"
	case codes.FailedPrecondition:
		return "FAILED_PRECONDITION"
	case codes.ResourceExhausted:
		return "RESOURCE_EXHAUSTED"
	case codes.Unavailable:
		return "SERVICE_UNAVAILABLE"
	case codes.DeadlineExceeded:
		return "DEADLINE_EXCEEDED"
	case codes.Aborted:
		return "CONFLICT"
	case codes.Unimplemented:
		return "NOT_IMPLEMENTED"
	default:
		return "INTERNAL"
	}
}
