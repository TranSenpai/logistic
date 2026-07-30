package biz

import "errors"

var (
	// 1. Lỗi chí mạng: Sai logic, sai data format -> Không bao giờ retry
	ErrNonRetryable = errors.New("fatal_error_do_not_retry")

	// 2. Lỗi quá tải/Mất kết nối: DB đang update, Service đang deploy -> Chờ 1 chút
	ErrRetryWithDelay = errors.New("service_unavailable_retry_later")
)
