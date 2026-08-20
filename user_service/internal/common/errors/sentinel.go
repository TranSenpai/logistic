// Package errors liệt kê TOÀN BỘ lỗi nghiệp vụ mà user_service có thể trả ra.
//
// Gom vào một chỗ để: (1) không có hai nơi cùng định nghĩa "user không tồn tại"
// với hai câu chữ khác nhau, (2) đọc file này là biết service có thể hỏng theo
// những cách nào, (3) mã lỗi (Code) trở thành hợp đồng ổn định với client —
// app mobile switch theo USER_NOT_FOUND chứ không phải theo câu tiếng Việt.
package errors

import "github.com/logistic/pkg/apperr"

var (
	// --- Tham số đầu vào ---
	ErrInvalidUserID    = apperr.InvalidArgument("INVALID_USER_ID", "user id không hợp lệ")
	ErrInvalidAddressID = apperr.InvalidArgument("INVALID_ADDRESS_ID", "address id không hợp lệ")
	ErrInvalidDeviceID  = apperr.InvalidArgument("INVALID_DEVICE_ID", "device id không hợp lệ")
	ErrInvalidRole      = apperr.InvalidArgument("INVALID_ROLE", "role phải là driver, shipper hoặc admin")
	ErrInvalidStatus    = apperr.InvalidArgument("INVALID_STATUS", "status phải là active, banned hoặc suspended")
	ErrInvalidKycStatus = apperr.InvalidArgument("INVALID_KYC_STATUS", "kyc_status phải là pending, approved hoặc rejected")
	ErrInvalidAddrType  = apperr.InvalidArgument("INVALID_ADDRESS_TYPE", "address_type phải là pickup, delivery hoặc both")
	ErrInvalidPlatform  = apperr.InvalidArgument("INVALID_PLATFORM", "platform phải là android, ios hoặc web")
	ErrPhoneRequired    = apperr.InvalidArgument("PHONE_REQUIRED", "số điện thoại là bắt buộc")
	ErrPasswordTooShort = apperr.InvalidArgument("PASSWORD_TOO_SHORT", "mật khẩu phải có ít nhất 6 ký tự")
	ErrAddressRequired  = apperr.InvalidArgument("ADDRESS_LINE_REQUIRED", "địa chỉ (line1) là bắt buộc")
	ErrDeviceTokenEmpty = apperr.InvalidArgument("DEVICE_TOKEN_REQUIRED", "device_token là bắt buộc")

	// --- Không tìm thấy ---
	ErrUserNotFound           = apperr.NotFound("USER_NOT_FOUND", "không tìm thấy người dùng")
	ErrDriverProfileNotFound  = apperr.NotFound("DRIVER_PROFILE_NOT_FOUND", "không tìm thấy hồ sơ tài xế")
	ErrShipperProfileNotFound = apperr.NotFound("SHIPPER_PROFILE_NOT_FOUND", "không tìm thấy hồ sơ chủ hàng")
	ErrAddressNotFound        = apperr.NotFound("ADDRESS_NOT_FOUND", "không tìm thấy địa chỉ")
	ErrDeviceNotFound         = apperr.NotFound("DEVICE_NOT_FOUND", "không tìm thấy thiết bị")

	// --- Trùng lặp ---
	ErrPhoneAlreadyUsed   = apperr.AlreadyExists("PHONE_ALREADY_USED", "số điện thoại đã được đăng ký")
	ErrEmailAlreadyUsed   = apperr.AlreadyExists("EMAIL_ALREADY_USED", "email đã được đăng ký")
	ErrLicenseAlreadyUsed = apperr.AlreadyExists("LICENSE_ALREADY_USED", "số bằng lái đã được dùng bởi tài xế khác")
	ErrIDCardAlreadyUsed  = apperr.AlreadyExists("ID_CARD_ALREADY_USED", "số CCCD đã được dùng bởi tài xế khác")

	// --- Sai điều kiện nghiệp vụ ---
	ErrNotADriver         = apperr.FailedPrecondition("NOT_A_DRIVER", "người dùng này không phải tài xế")
	ErrNotAShipper        = apperr.FailedPrecondition("NOT_A_SHIPPER", "người dùng này không phải chủ hàng")
	ErrKycAlreadyReviewed = apperr.FailedPrecondition("KYC_ALREADY_REVIEWED", "hồ sơ KYC đã được duyệt trước đó")
	ErrUserBanned         = apperr.FailedPrecondition("USER_BANNED", "tài khoản đang bị khoá")
	ErrKycNotApproved     = apperr.FailedPrecondition("KYC_NOT_APPROVED", "hồ sơ KYC chưa được duyệt")

	// --- Quyền ---
	ErrAddressNotOwned = apperr.PermissionDenied("ADDRESS_NOT_OWNED", "địa chỉ không thuộc về người dùng này")
	ErrDeviceNotOwned  = apperr.PermissionDenied("DEVICE_NOT_OWNED", "thiết bị không thuộc về người dùng này")

	// --- Hạ tầng ---
	ErrDatabase = apperr.Internal("DATABASE_ERROR", "lỗi truy cập cơ sở dữ liệu")
)
