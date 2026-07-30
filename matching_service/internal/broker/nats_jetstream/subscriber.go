package nats_jetstream

import (
	"context"
	"errors"
	"log"
	"matching_service/internal/biz"
	"time"

	"github.com/nats-io/nats.go"
)

type natsSubcriber struct {
	natJetStreamContext nats.JetStreamContext
}

func InitSubcriber(natJetStreamContext nats.JetStreamContext) biz.EventConsumer {
	return &natsSubcriber{
		natJetStreamContext: natJetStreamContext,
	}
}

var _ biz.EventConsumer = (*natsSubcriber)(nil)

func (n *natsSubcriber) Consume(ctx context.Context, topic string, handler func(bucket []byte) error) error {
	_, err := n.natJetStreamContext.Subscribe(topic, func(msg *nats.Msg) {
		err := handler(msg.Data)
		if err != nil {
			if errors.Is(err, biz.ErrNonRetryable) {
				log.Printf("Discarding poison pill message: %v", err)
				msg.Ack()
				return
			}

			if errors.Is(err, biz.ErrRetryWithDelay) {
				log.Printf("Service busy, retrying in 5s: %v", err)
				msg.NakWithDelay(5 * time.Second)
				return
			}

			log.Printf("Transient error, retrying immediately: %v", err)
			msg.Nak()
			return

		}
		msg.Ack()

	}, nats.ManualAck())

	return err
}
