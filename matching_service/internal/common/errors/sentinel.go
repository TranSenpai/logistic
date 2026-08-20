// Package errors là bảng mã lỗi của matching_service.
//
// Trước đây các sentinel ở đây là errors.New thuần, nên khi lỗi đi ra khỏi
// service, gRPC gán cho tất cả cùng một mã codes.Unknown — client không phân
// biệt được "không tìm thấy đơn" với "hết tiền trong ví".
//
// Nay chúng là *apperr.Error, mang sẵn Kind và Code. ErrorInterceptor trong
// pkg/middleware đọc hai thứ đó để trả về đúng NOT_FOUND / FAILED_PRECONDITION...
// Cách dùng cũ vẫn chạy nguyên vẹn: fmt.Errorf("%w: ...", ErrInvalidID) vẫn bọc
// được, và errors.Is / apperr.From đều xuyên qua lớp bọc đó.
package errors

import "github.com/logistic/pkg/apperr"

var (
	// General
	ErrInvalidID = apperr.InvalidArgument("INVALID_ID", "định dạng id không hợp lệ")

	// Repo
	ErrRecordNotFound  = apperr.NotFound("RECORD_NOT_FOUND", "không tìm thấy bản ghi")
	ErrInlavidInput    = apperr.InvalidArgument("INVALID_INPUT", "dữ liệu đầu vào không hợp lệ")
	ErrDuplicateRecord = apperr.AlreadyExists("DUPLICATE_RECORD", "bản ghi đã tồn tại")

	// Biz
	ErrValidationFailed    = apperr.InvalidArgument("VALIDATION_FAILED", "dữ liệu không hợp lệ")
	ErrUnauthorized        = apperr.PermissionDenied("UNAUTHORIZED", "không có quyền thực hiện hành động này")
	ErrInsufficientBalance = apperr.FailedPrecondition("INSUFFICIENT_BALANCE", "số dư ví không đủ để đặt cọc")

	// Trạng thái nghiệp vụ của vòng thương lượng
	ErrBidNotPending     = apperr.FailedPrecondition("BID_NOT_PENDING", "đơn hàng không còn ở trạng thái chờ")
	ErrBidNotNegotiating = apperr.FailedPrecondition("BID_NOT_NEGOTIATING", "đơn hàng không ở trạng thái đang thương lượng")

	// Internal
	ErrInternalServer = apperr.Internal("INTERNAL_SERVER_ERROR", "lỗi hệ thống")
)
