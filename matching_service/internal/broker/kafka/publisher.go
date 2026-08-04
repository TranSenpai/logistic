package kafka

import (
	"context"
	"errors"
	"fmt"
	"log"
	"matching_service/internal/biz"
	"matching_service/internal/broker"
	"matching_service/internal/entity"
	"matching_service/internal/mapper"

	pb "github.com/logistic/api/logistic/matching_service/v1"

	"github.com/IBM/sarama"
	"google.golang.org/protobuf/proto"
)

type kafkaPublisher struct {
	producer sarama.SyncProducer
	mapper   mapper.AppMapper
}

var _ biz.EventPublisher = (*kafkaPublisher)(nil)

func NewKafkaPublisher(brokers []string, appMapper mapper.AppMapper) (biz.EventPublisher, error) {
	config := sarama.NewConfig()

	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &kafkaPublisher{producer: producer, mapper: appMapper}, err
}

func (p *kafkaPublisher) payloadToBytes(payload any) ([]byte, error) {
	switch v := payload.(type) {
	case []byte:
		return v, nil
	case *entity.Bid:
		if v == nil {
			return nil, broker.ErrNilMessage
		}
		pbBid := p.mapper.EntityBidToPbBid(*v)

		return proto.Marshal(pbBid)
	case []entity.Ask:
		if v == nil {
			return nil, broker.ErrNilMessage
		}
		payload := &pb.Asks{Asks: p.mapper.EntityAskListToPbAskList(v)}

		return proto.Marshal(payload)
	default:
		return nil, fmt.Errorf("unsupported payload type for protobuf serialization")
	}
}

func (p *kafkaPublisher) Publish(ctx context.Context, msg *biz.EventMessage) error {
	if msg == nil {
		return broker.ErrNilMessage
	}

	payloadBytes, err := p.payloadToBytes(msg.Payload)
	if err != nil {
		return err
	}

	kafkaMgs := sarama.ProducerMessage{
		Topic: msg.Topic,
		Value: sarama.ByteEncoder(payloadBytes),
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
