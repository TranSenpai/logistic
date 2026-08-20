package repo

import (
	"context"
	"errors"
	"strings"

	"user_service/ent"
	cerr "user_service/internal/common/errors"

	"github.com/logistic/pkg/apperr"
)

// wrapError dịch lỗi của ent/Postgres sang mã lỗi nghiệp vụ.
//
// Đây là RANH GIỚI: từ đây trở lên (biz, controller) không ai còn phải biết
// *ent.NotFoundError hay chuỗi "duplicate key value violates unique constraint"
// trông như thế nào. Nếu không có lớp dịch này, một lỗi trùng số điện thoại sẽ
// đi thẳng ra client thành HTTP 500 kèm nguyên văn câu SQL của Postgres.
//
// notFound: mã lỗi trả về khi ent báo không tìm thấy — mỗi bảng một mã riêng nên
// caller truyền vào, ví dụ ErrUserNotFound hoặc ErrAddressNotFound.
func wrapError(err error, notFound *apperr.Error) error {
	if err == nil {
		return nil
	}

	// Context bị huỷ (client bỏ request, deadline hết) không phải lỗi của ta.
	if errors.Is(err, context.Canceled) {
		return apperr.New(apperr.KindTimeout, "REQUEST_CANCELLED", "yêu cầu đã bị huỷ").WithCause(err)
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return apperr.New(apperr.KindTimeout, "REQUEST_TIMEOUT", "yêu cầu quá thời gian chờ").WithCause(err)
	}

	if ent.IsNotFound(err) {
		if notFound == nil {
			notFound = cerr.ErrUserNotFound
		}
		return notFound.WithCause(err)
	}

	if ent.IsConstraintError(err) {
		return mapConstraint(err)
	}

	if ent.IsValidationError(err) {
		return apperr.InvalidArgument("VALIDATION_FAILED", "dữ liệu không hợp lệ").WithCause(err)
	}

	if ent.IsNotSingular(err) {
		return apperr.Conflict("NOT_SINGULAR", "truy vấn trả về nhiều hơn một bản ghi").WithCause(err)
	}

	return cerr.ErrDatabase.WithCause(err)
}

// mapConstraint đọc tên constraint trong thông báo lỗi để biết CỘT NÀO bị trùng.
// Postgres nhét tên index vào message (vd: "users_phone_key"), nên khớp chuỗi là
// cách duy nhất mà không phải cài thêm driver-specific error parsing.
func mapConstraint(err error) error {
	msg := strings.ToLower(err.Error())

	switch {
	case strings.Contains(msg, "phone"):
		return cerr.ErrPhoneAlreadyUsed.WithCause(err)
	case strings.Contains(msg, "email"):
		return cerr.ErrEmailAlreadyUsed.WithCause(err)
	case strings.Contains(msg, "license_number"):
		return cerr.ErrLicenseAlreadyUsed.WithCause(err)
	case strings.Contains(msg, "id_card"):
		return cerr.ErrIDCardAlreadyUsed.WithCause(err)
	case strings.Contains(msg, "device_token"):
		return apperr.AlreadyExists("DEVICE_TOKEN_EXISTS", "thiết bị này đã được đăng ký").WithCause(err)
	default:
		return apperr.Conflict("CONSTRAINT_VIOLATION", "dữ liệu vi phạm ràng buộc của hệ thống").WithCause(err)
	}
}
