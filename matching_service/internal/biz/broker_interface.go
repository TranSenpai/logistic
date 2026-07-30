package biz

import (
	"context"
)

type EventMessage struct {
	Header  string
	Topic   string
	Payload []byte
}

type EventPublisher interface {
	Publish(ctx context.Context, msg *EventMessage) error
}

type EventConsumer interface {
	Consume(ctx context.Context, topic string, handler func(bucket []byte) error) error
}
