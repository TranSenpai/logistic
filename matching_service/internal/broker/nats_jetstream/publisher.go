package nats_jetstream

import (
	"context"
	"errors"
	"fmt"
	"log"
	"matching_service/internal/biz"

	"github.com/nats-io/nats.go"
)

var _ biz.EventPublisher = (*natsPublisher)(nil)

type natsPublisher struct {
	natJetStreamContext nats.JetStreamContext
}

func InitPublisher(natJetStreamContext nats.JetStreamContext) biz.EventPublisher {
	return &natsPublisher{
		natJetStreamContext: natJetStreamContext,
	}
}

func (n *natsPublisher) Publish(ctx context.Context, msg *biz.EventMessage) error {
	ack, err := n.natJetStreamContext.Publish(msg.Topic, msg.Payload)
	if err != nil {
		if errors.Is(err, nats.ErrTimeout) || errors.Is(err, nats.ErrConnectionClosed) {
			log.Printf("NATS connection error, retry requested: %v", err)
			return fmt.Errorf("%w: %v", biz.ErrRetryWithDelay, err)
		}

		if errors.Is(err, nats.ErrMaxPayload) || errors.Is(err, nats.ErrNoResponders) {
			log.Printf("Configuration error or payload too large, do not retry: %v", err)
			return fmt.Errorf("%w: %v", biz.ErrNonRetryable, err)
		}

		return fmt.Errorf("%w: unknown NATS error: %v", biz.ErrRetryWithDelay, err)
	}

	log.Printf("Message published successfully! Stream: %s, Sequence (ID): %d, Duplicate: %v",
		ack.Stream, ack.Sequence, ack.Duplicate)

	return err
}
