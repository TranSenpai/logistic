package nats_jetstream

import (
	"context"
	"errors"
	"fmt"
	"log"
	"matching_service/internal/biz"
	"matching_service/internal/broker"
	"matching_service/internal/mapper"

	"github.com/nats-io/nats.go"
)

var _ biz.EventPublisher = (*natsPublisher)(nil)

type natsPublisher struct {
	natJetStreamContext nats.JetStreamContext
	mapper              mapper.AppMapper
}

func InitPublisher(natJetStreamContext nats.JetStreamContext, appMapper mapper.AppMapper) biz.EventPublisher {
	return &natsPublisher{
		natJetStreamContext: natJetStreamContext,
		mapper:              appMapper,
	}
}

func (n *natsPublisher) payloadToBytes(payload any) ([]byte, error) {
	switch v := payload.(type) {
	case []byte:
		return v, nil
	// TƯƠNG LAI BỔ SUNG TYPE ASSERTION Ở ĐÂY:
	// case *entity.Bid:
	//     pbBid := n.mapper.EntityBidToPbBid(v) 
	//     return proto.Marshal(pbBid)
	default:
		return nil, fmt.Errorf("unsupported payload type for protobuf serialization")
	}
}

func (n *natsPublisher) Publish(ctx context.Context, msg *biz.EventMessage) error {
	if msg == nil {
		return broker.ErrNilMessage
	}

	payloadBytes, err := n.payloadToBytes(msg.Payload)
	if err != nil {
		return err
	}

	ack, err := n.natJetStreamContext.Publish(msg.Topic, payloadBytes)
	if err != nil {
		return parseNatsError(err)
	}

	log.Printf("Message published successfully! Stream: %s, Sequence (ID): %d, Duplicate: %v",
		ack.Stream, ack.Sequence, ack.Duplicate)

	return nil
}

// parseNatsError phân loại các lỗi đặc thù của NATS thành các lỗi Biz chuẩn
func parseNatsError(err error) error {
	// 1. Nhóm lỗi mạng, Timeout, Đứt kết nối -> Có thể thử lại sau (Retry)
	if errors.Is(err, nats.ErrTimeout) ||
		errors.Is(err, nats.ErrConnectionClosed) ||
		errors.Is(err, nats.ErrNoServers) ||
		errors.Is(err, nats.ErrStaleConnection) ||
		errors.Is(err, nats.ErrDisconnected) {
		log.Printf("NATS network/timeout error, retry requested: %v", err)
		return fmt.Errorf("%w: %v", biz.ErrRetryWithDelay, err)
	}

	// 2. Nhóm lỗi logic, Cấu hình, Quá tải -> Tuyệt đối cấm Retry (Sẽ kẹt hệ thống)
	// (Ví dụ: Sai Topic, Gói tin to hơn mức cho phép, Lỗi xác thực, Stream bị xóa...)
	if errors.Is(err, nats.ErrMaxPayload) ||
		errors.Is(err, nats.ErrNoResponders) ||
		errors.Is(err, nats.ErrBadSubject) ||
		errors.Is(err, nats.ErrAuthorization) {
		log.Printf("NATS configuration/payload error, do not retry: %v", err)
		return fmt.Errorf("%w: %v", biz.ErrNonRetryable, err)
	}

	// 3. JetStream API Errors (Lỗi đặc thù của JetStream)
	// Các lỗi JetStream thường trả về kiểu *nats.APIError, ta check theo error string hoặc type
	if apiErr, ok := errors.AsType[*nats.APIError](err); ok {
		log.Printf("NATS JetStream API Error, do not retry: %v", apiErr)
		return fmt.Errorf("%w: %v", biz.ErrNonRetryable, apiErr)
	}

	// Mặc định: Trả về lỗi yêu cầu Retry để an toàn
	return fmt.Errorf("%w: unknown NATS error: %v", biz.ErrRetryWithDelay, err)
}
