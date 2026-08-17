package biz

import (
	"context"
	"errors"
)

type Header struct {
	Key   []byte
	Value []byte
}

type EventMessage struct {
	Header  *Header
	Topic   string
	Key     string
	Payload any
}

type EventPublisher interface {
	Publish(ctx context.Context, msg *EventMessage) error
}

type EventConsumer interface {
	Consume(ctx context.Context, topic string, handler func(ctx context.Context, bucket []byte) error) error
}

// Sentinel errors để Consumer phân biệt lỗi nghiệp vụ vs lỗi hệ thống
var (
	// ErrNonRetryable: Lỗi nghiệp vụ, không cần retry (ví dụ: không đủ số dư)
	ErrNonRetryable = errors.New("non-retryable business error")

	// ErrRetryWithDelay: Lỗi tạm thời, cần retry sau
	ErrRetryWithDelay = errors.New("retryable error, retry with delay")

	// ErrServiceUnavailable503: Service bị down, không thể xử lý
	ErrServiceUnavailable503 = errors.New("service unavailable")
)
