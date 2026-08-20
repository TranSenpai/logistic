// Package rabbitmq cài đặt biz.Notifier bằng RabbitMQ topic exchange.
//
// Đây là đầu PHÁT của luồng thông báo; đầu nhận nằm ở
// notification_service/internal/consumer. Hợp đồng dữ liệu giữa hai bên là
// package github.com/logistic/pkg/events, nên hai service không thể lệch nhau
// về tên trường mà vẫn biên dịch được.
package rabbitmq

import (
	"context"
	"fmt"
	"log"
	"time"

	"matching_service/internal/biz"
	"matching_service/internal/entity"

	"github.com/google/uuid"
	"github.com/logistic/pkg/events"
	"github.com/logistic/pkg/mq"
)

type notifier struct {
	pub    *mq.Publisher
	source string
}

var _ biz.Notifier = (*notifier)(nil)

func NewNotifier(pub *mq.Publisher) biz.Notifier {
	return &notifier{pub: pub, source: "matching_service"}
}

// publish gói payload vào Envelope rồi đẩy đi.
//
// event_id sinh mới cho MỖI lần phát và trở thành MessageId của message.
// notification_service dùng chính id này làm khoá chống trùng, nên hai lần
// publish khác nhau của cùng một nghiệp vụ sẽ được coi là hai sự kiện — đúng
// như mong muốn, còn một message được broker giao lại thì không.
func (n *notifier) publish(ctx context.Context, routingKey string, data map[string]any) error {
	env := events.Envelope{
		EventID:    uuid.Must(uuid.NewV7()).String(),
		EventType:  routingKey,
		OccurredAt: time.Now().UTC(),
		Source:     n.source,
		Data:       data,
	}

	if err := n.pub.Publish(ctx, routingKey, env.EventID, env); err != nil {
		return fmt.Errorf("phát sự kiện %s thất bại: %w", routingKey, err)
	}

	log.Printf("[notifier] đã phát %s (event_id=%s)", routingKey, env.EventID)
	return nil
}

// NotifyDriverCandidates — luồng "chủ hàng tìm xe".
//
// engine đã lọc ra các ask phù hợp; ta quy chúng thành danh sách ứng viên kèm
// khoảng cách để notification_service dựng câu chữ "cách bạn X km".
func (n *notifier) NotifyDriverCandidates(ctx context.Context, bid *entity.Bid, asks []entity.Ask) error {
	if bid == nil || len(asks) == 0 {
		return nil
	}

	candidates := make([]events.DriverCandidate, 0, len(asks))
	for _, ask := range asks {
		candidates = append(candidates, events.DriverCandidate{
			DriverID:  ask.DriverID.String(),
			AskID:     ask.ID.String(),
			VehicleID: ask.VehicleID.String(),
			Phone:     ask.DriverPhone,
			Email:     ask.DriverMail,
			DistanceKm: haversineKm(
				bid.Origin.Latitude, bid.Origin.Longitude,
				ask.CurrentLocation.Latitude, ask.CurrentLocation.Longitude,
			),
		})
	}

	payload := events.DriverCandidatesFound{
		BidID:          bid.ID.String(),
		ShipperID:      bid.ShipperID.String(),
		OriginZoneID:   bid.Origin.ZoneID,
		OriginLat:      bid.Origin.Latitude,
		OriginLng:      bid.Origin.Longitude,
		DestinationLat: bid.Destination.Latitude,
		DestinationLng: bid.Destination.Longitude,
		WeightKg:       bid.WeightKg,
		VolumeM3:       bid.VolumeM3,
		MaxPrice:       bid.MaxPrice,
		Candidates:     candidates,
	}

	return n.publish(ctx, events.RoutingKeyDriverCandidatesFound, toMap(payload))
}

// NotifyMatchFound — luồng "đã tìm được xe".
func (n *notifier) NotifyMatchFound(ctx context.Context, contract *entity.MatchContract, bid *entity.Bid, ask *entity.Ask) error {
	if contract == nil || bid == nil || ask == nil {
		return nil
	}

	payload := events.MatchFound{
		ContractID:       contract.ID.String(),
		BidID:            bid.ID.String(),
		AskID:            ask.ID.String(),
		ShipperID:        bid.ShipperID.String(),
		DriverID:         ask.DriverID.String(),
		VehicleID:        ask.VehicleID.String(),
		ConsensusPrice:   contract.ConsensusPrice,
		ConsensusDeposit: contract.ConsensusDeposit,
		ShipperPhone:     bid.ShipperPhone,
		ShipperEmail:     bid.ShipperMail,
		DriverPhone:      ask.DriverPhone,
		DriverEmail:      ask.DriverMail,
	}

	return n.publish(ctx, events.RoutingKeyMatchFound, toMap(payload))
}

func (n *notifier) NotifyOfferReceived(ctx context.Context, bid *entity.Bid, ask *entity.Ask, price float64) error {
	if bid == nil || ask == nil {
		return nil
	}

	payload := events.OfferReceived{
		BidID:     bid.ID.String(),
		AskID:     ask.ID.String(),
		ShipperID: bid.ShipperID.String(),
		DriverID:  ask.DriverID.String(),
		Price:     price,
	}

	return n.publish(ctx, events.RoutingKeyOfferReceived, toMap(payload))
}

func (n *notifier) NotifyOfferRejected(ctx context.Context, bid *entity.Bid, ask *entity.Ask, reason string) error {
	if bid == nil || ask == nil {
		return nil
	}

	payload := events.OfferRejected{
		BidID:    bid.ID.String(),
		AskID:    ask.ID.String(),
		DriverID: ask.DriverID.String(),
		Reason:   reason,
	}

	return n.publish(ctx, events.RoutingKeyOfferRejected, toMap(payload))
}

func (n *notifier) NotifyCargoSuggested(ctx context.Context, ask *entity.Ask, bids []entity.Bid) error {
	if ask == nil || len(bids) == 0 {
		return nil
	}

	bidIDs := make([]string, 0, len(bids))
	for _, b := range bids {
		bidIDs = append(bidIDs, b.ID.String())
	}

	payload := events.CargoSuggested{
		AskID:      ask.ID.String(),
		DriverID:   ask.DriverID.String(),
		VehicleID:  ask.VehicleID.String(),
		BidIDs:     bidIDs,
		TotalFound: len(bids),
	}

	return n.publish(ctx, events.RoutingKeyCargoSuggested, toMap(payload))
}
