package biz

import (
	"context"
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
