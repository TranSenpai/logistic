package repo

import (
	"context"
	"errors"
	"strings"

	"notification_service/ent"
	cerr "notification_service/internal/common/errors"

	"github.com/logistic/pkg/apperr"
)

// wrapError dịch lỗi ent/Postgres sang mã lỗi nghiệp vụ.
func wrapError(err error, notFound *apperr.Error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, context.Canceled) {
		return apperr.New(apperr.KindTimeout, "REQUEST_CANCELLED", "yêu cầu đã bị huỷ").WithCause(err)
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return apperr.New(apperr.KindTimeout, "REQUEST_TIMEOUT", "yêu cầu quá thời gian chờ").WithCause(err)
	}

	if ent.IsNotFound(err) {
		if notFound == nil {
			notFound = cerr.ErrNotificationNotFound
		}
		return notFound.WithCause(err)
	}

	if ent.IsConstraintError(err) {
		msg := strings.ToLower(err.Error())
		switch {
		case strings.Contains(msg, "event_id"):
			// Đây KHÔNG phải lỗi thật: consumer dùng nó để nhận ra message trùng.
			return ErrDuplicateEvent.WithCause(err)
		case strings.Contains(msg, "code"):
			return cerr.ErrTemplateCodeExists.WithCause(err)
		case strings.Contains(msg, "user_id"):
			return apperr.AlreadyExists("PREFERENCE_EXISTS", "người dùng đã có cài đặt thông báo").WithCause(err)
		default:
			return apperr.Conflict("CONSTRAINT_VIOLATION", "dữ liệu vi phạm ràng buộc của hệ thống").WithCause(err)
		}
	}

	if ent.IsValidationError(err) {
		return apperr.InvalidArgument("VALIDATION_FAILED", "dữ liệu không hợp lệ").WithCause(err)
	}
	if ent.IsNotSingular(err) {
		return apperr.Conflict("NOT_SINGULAR", "truy vấn trả về nhiều hơn một bản ghi").WithCause(err)
	}

	return cerr.ErrDatabase.WithCause(err)
}

// ErrDuplicateEvent báo rằng event_id đã được xử lý trước đó.
//
// Đây là tín hiệu BÌNH THƯỜNG trong hệ thống at-least-once, không phải sự cố:
// consumer bắt được nó thì ACK message rồi đi tiếp, thay vì retry vô ích.
var ErrDuplicateEvent = apperr.AlreadyExists("EVENT_ALREADY_PROCESSED", "sự kiện này đã được xử lý")
