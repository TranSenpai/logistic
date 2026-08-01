package kafka

import (
	"context"
	"errors"
	"fmt"
	"log"
	"matching_service/internal/biz"
	"matching_service/internal/broker"

	"github.com/IBM/sarama"
)

type kafkaPublisher struct {
	producer sarama.SyncProducer
}

var _ biz.EventPublisher = (*kafkaPublisher)(nil)

func NewKafkaPublisher(brokers []string) (biz.EventPublisher, error) {
	config := sarama.NewConfig()

	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &kafkaPublisher{producer: producer}, err
}

func (p *kafkaPublisher) Publish(ctx context.Context, msg *biz.EventMessage) error {
	if msg == nil {
		return broker.ErrNilMessage
	}

	kafkaMgs := sarama.ProducerMessage{
		Topic: msg.Topic,
		Value: sarama.ByteEncoder(msg.Payload),
	}
	if msg.Key != "" {
		kafkaMgs.Key = sarama.StringEncoder(msg.Key)
	}

	if msg.Header != nil {
		kafkaMgs.Headers = []sarama.RecordHeader{
			{
				Key:   msg.Header.Key,
				Value: msg.Header.Value,
			},
		}
	}

	partition, offset, err := p.producer.SendMessage(&kafkaMgs)
	if err != nil {
		// Leverage Sarama's KError type
		if kerr, ok := errors.AsType[sarama.KError](err); ok {
			switch kerr {
			case sarama.ErrRequestTimedOut,
				sarama.ErrNetworkException,
				sarama.ErrBrokerNotAvailable,
				sarama.ErrNotLeaderForPartition:
				log.Printf("Kafka connection/timeout error, retry requested: %v", kerr)
				return fmt.Errorf("%w: %v", biz.ErrRetryWithDelay, kerr)

			case sarama.ErrMessageSizeTooLarge,
				sarama.ErrInvalidMessage,
				sarama.ErrUnknownTopicOrPartition:
				log.Printf("Configuration error or payload too large, do not retry: %v", kerr)
				return fmt.Errorf("%w: %v", biz.ErrNonRetryable, kerr)
			}
		}

		// Fallback for any unhandled Sarama errors or non-KError issues -> Return 503
		log.Printf("Unknown Kafka error, returning 503: %v", err)
		return fmt.Errorf("%w: %v", biz.ErrServiceUnavailable503, err)
	}
	log.Printf("Message published successfully! Partition: %d, Offset: %d", partition, offset)

	return err
}
