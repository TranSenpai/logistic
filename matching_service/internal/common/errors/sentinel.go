package errors

import "github.com/logistic/pkg/apperr"

var (
	ErrInvalidID = apperr.InvalidArgument("INVALID_ID", "định dạng id không hợp lệ")

	ErrRecordNotFound  = apperr.NotFound("RECORD_NOT_FOUND", "không tìm thấy bản ghi")
	ErrInlavidInput    = apperr.InvalidArgument("INVALID_INPUT", "dữ liệu đầu vào không hợp lệ")
	ErrDuplicateRecord = apperr.AlreadyExists("DUPLICATE_RECORD", "bản ghi đã tồn tại")

	ErrValidationFailed    = apperr.InvalidArgument("VALIDATION_FAILED", "dữ liệu không hợp lệ")
	ErrUnauthorized        = apperr.PermissionDenied("UNAUTHORIZED", "không có quyền thực hiện hành động này")
	ErrInsufficientBalance = apperr.FailedPrecondition("INSUFFICIENT_BALANCE", "số dư ví không đủ để đặt cọc")

	ErrBidNotPending     = apperr.FailedPrecondition("BID_NOT_PENDING", "đơn hàng không còn ở trạng thái chờ")
	ErrBidNotNegotiating = apperr.FailedPrecondition("BID_NOT_NEGOTIATING", "đơn hàng không ở trạng thái đang thương lượng")

	ErrInternalServer = apperr.Internal("INTERNAL_SERVER_ERROR", "lỗi hệ thống")
)