package entity

import "errors"

// Sentinel errors dùng chung cho toàn bộ hệ thống.
// Đặt ở tầng entity để cả tầng Repo và Biz đều có thể gọi mà không bị cyclic import.
var (
	// Lỗi nil
	ErrNilBid   = errors.New("bid is nil")
	ErrNilAsk   = errors.New("ask is nil")
	ErrNilMatch = errors.New("match is nil")

	// Lỗi liên quan đến spatial
	ErrEmptyLocation   = errors.New("empty location")
	ErrInvalidLocation = errors.New("invalid location")
	ErrEmptyZoneID     = errors.New("empty zone ID")

	// Lỗi liên quan đến tìm kiếm
	ErrBidNotFound   = errors.New("bid not found")
	ErrAskNotFound   = errors.New("ask not found")
	ErrMatchNotFound = errors.New("match not found")

	// Lỗi liên quan đến trạng thái và ghép đơn
	ErrAlreadyMatched  = errors.New("entity is already matched")
	ErrInvalidStatus   = errors.New("invalid status transition")
	ErrNotEnoughVolume = errors.New("not enough volume available")
	ErrNotEnoughWeight = errors.New("not enough weight available")
	ErrPriceMismatch   = errors.New("price conditions not met")

	// Lỗi hệ thống
	ErrInternal = errors.New("internal system error")
)
