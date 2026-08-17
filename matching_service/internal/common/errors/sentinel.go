package errors

import "errors"

// Các Sentinel Errors chuẩn dùng chung cho toàn dự án
// Được định nghĩa ở common để cả tầng Repo và Middleware đều có thể import mà không bị Cyclic Dependency.

var (
	// General error
	ErrInvalidID = errors.New("invalid id format")

	//  Repo
	ErrRecordNotFound  = errors.New("record not found")
	ErrInlavidInput    = errors.New("invalid input")
	ErrDuplicateRecord = errors.New("duplicate record")

	//  Biz
	ErrValidationFailed    = errors.New("validation failed")
	ErrUnauthorized        = errors.New("unauthorized action")
	ErrInsufficientBalance = errors.New("insufficient balance")

	// Internal Server Error
	ErrInternalServer = errors.New("internal server error")
)
