package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"

	"notification_service/internal/biz"
	"notification_service/internal/entity"
	"notification_service/internal/repo"

	"github.com/google/uuid"
	"github.com/logistic/pkg/apperr"
	"github.com/logistic/pkg/events"
	"github.com/logistic/pkg/mq"
)

type MatchingConsumer struct {
	engine biz.NotificationEngine
}

func NewMatchingConsumer(engine biz.NotificationEngine) *MatchingConsumer {
	return &MatchingConsumer{engine: engine}
}

func (c *MatchingConsumer) Handle(ctx context.Context, d mq.Delivery) error {
	var env events.Envelope
	if err := json.Unmarshal(d.Body, &env); err != nil {
		log.Printf("[consumer] bỏ qua message không parse được (routing=%s id=%s): %v",
			d.RoutingKey, d.MessageID, err)
		return nil
	}

	eventID := env.EventID
	if eventID == "" {
		eventID = d.MessageID
	}
	if eventID == "" {
		log.Printf("[consumer] bỏ qua message không có event_id (routing=%s)", d.RoutingKey)
		return nil
	}

	params, err := c.buildNotifications(d.RoutingKey, &env)
	if err != nil {
		log.Printf("[consumer] không dựng được thông báo cho %s (id=%s): %v", d.RoutingKey, eventID, err)
		return nil
	}
	if len(params) == 0 {
		log.Printf("[consumer] %s (id=%s) không có người nhận hợp lệ", d.RoutingKey, eventID)
		return nil
	}

	count, err := c.engine.DispatchEvent(ctx, eventID, d.RoutingKey, env.Source, params)
	if err != nil {
		if errors.Is(err, repo.ErrDuplicateEvent) || isDuplicate(err) {
			log.Printf("[consumer] event %s đã xử lý trước đó — bỏ qua", eventID)
			return nil
		}

		return fmt.Errorf("dispatch event %s: %w", eventID, err)
	}

	log.Printf("[consumer] %s (id=%s) -> đã tạo %d thông báo", d.RoutingKey, eventID, count)
	return nil
}

func isDuplicate(err error) bool {
	appErr, ok := apperr.From(err)
	return ok && appErr.Code == "EVENT_ALREADY_PROCESSED"
}

func (c *MatchingConsumer) buildNotifications(routingKey string, env *events.Envelope) ([]entity.CreateNotificationParam, error) {
	switch routingKey {
	case events.RoutingKeyDriverCandidatesFound:
		return c.onDriverCandidatesFound(env)
	case events.RoutingKeyMatchFound:
		return c.onMatchFound(env)
	case events.RoutingKeyOfferReceived:
		return c.onOfferReceived(env)
	case events.RoutingKeyOfferRejected:
		return c.onOfferRejected(env)
	case events.RoutingKeyCargoSuggested:
		return c.onCargoSuggested(env)
	default:

		return nil, nil
	}
}

func (c *MatchingConsumer) onDriverCandidatesFound(env *events.Envelope) ([]entity.CreateNotificationParam, error) {
	var payload events.DriverCandidatesFound
	if err := decodeData(env.Data, &payload); err != nil {
		return nil, err
	}

	params := make([]entity.CreateNotificationParam, 0, len(payload.Candidates))
	for _, cand := range payload.Candidates {
		driverID, err := uuid.Parse(cand.DriverID)
		if err != nil {
			log.Printf("[consumer] bỏ qua ứng viên có driver_id không hợp lệ: %q", cand.DriverID)
			continue
		}

		title := "Có đơn hàng phù hợp gần bạn"
		body := fmt.Sprintf(
			"Đơn hàng %.0f kg / %.1f m³ cách bạn %.1f km. Giá tối đa %.0f đ. Vào xem ngay để báo giá.",
			payload.WeightKg, payload.VolumeM3, cand.DistanceKm, payload.MaxPrice,
		)

		params = append(params, entity.CreateNotificationParam{
			UserID:        driverID,
			RecipientRole: entity.RoleDriver,
			Type:          entity.TypeDriverCandidate,
			Channel:       entity.ChannelPush,
			Title:         title,
			Body:          body,
			RefType:       entity.RefTypeBid,
			RefID:         payload.BidID,
			Data: biz.MarshalData(map[string]string{
				"bid_id":      payload.BidID,
				"ask_id":      cand.AskID,
				"vehicle_id":  cand.VehicleID,
				"distance_km": strconv.FormatFloat(cand.DistanceKm, 'f', 2, 64),
				"screen":      "BidDetail",
			}),
		})
	}

	return params, nil
}

func (c *MatchingConsumer) onMatchFound(env *events.Envelope) ([]entity.CreateNotificationParam, error) {
	var payload events.MatchFound
	if err := decodeData(env.Data, &payload); err != nil {
		return nil, err
	}

	data := biz.MarshalData(map[string]string{
		"contract_id": payload.ContractID,
		"bid_id":      payload.BidID,
		"ask_id":      payload.AskID,
		"vehicle_id":  payload.VehicleID,
		"screen":      "MatchDetail",
	})

	params := make([]entity.CreateNotificationParam, 0, 2)

	if shipperID, err := uuid.Parse(payload.ShipperID); err == nil {
		params = append(params, entity.CreateNotificationParam{
			UserID:        shipperID,
			RecipientRole: entity.RoleShipper,
			Type:          entity.TypeMatchFound,
			Channel:       entity.ChannelPush,
			Title:         "Đã tìm được xe cho đơn hàng của bạn",
			Body: fmt.Sprintf(
				"Đơn hàng của bạn đã được ghép với một tài xế. Giá chốt %.0f đ, đặt cọc %.0f đ.",
				payload.ConsensusPrice, payload.ConsensusDeposit,
			),
			RefType: entity.RefTypeMatch,
			RefID:   payload.ContractID,
			Data:    data,
		})
	} else {
		log.Printf("[consumer] match %s có shipper_id không hợp lệ: %q", payload.ContractID, payload.ShipperID)
	}

	if driverID, err := uuid.Parse(payload.DriverID); err == nil {
		params = append(params, entity.CreateNotificationParam{
			UserID:        driverID,
			RecipientRole: entity.RoleDriver,
			Type:          entity.TypeMatchFound,
			Channel:       entity.ChannelPush,
			Title:         "Bạn vừa nhận được một đơn hàng",
			Body: fmt.Sprintf(
				"Chuyến hàng đã được xác nhận. Giá chốt %.0f đ. Mở app để xem điểm lấy hàng.",
				payload.ConsensusPrice,
			),
			RefType: entity.RefTypeMatch,
			RefID:   payload.ContractID,
			Data:    data,
		})
	} else {
		log.Printf("[consumer] match %s có driver_id không hợp lệ: %q", payload.ContractID, payload.DriverID)
	}

	return params, nil
}

func (c *MatchingConsumer) onOfferReceived(env *events.Envelope) ([]entity.CreateNotificationParam, error) {
	var payload events.OfferReceived
	if err := decodeData(env.Data, &payload); err != nil {
		return nil, err
	}

	shipperID, err := uuid.Parse(payload.ShipperID)
	if err != nil {
		return nil, fmt.Errorf("shipper_id không hợp lệ: %w", err)
	}

	return []entity.CreateNotificationParam{{
		UserID:        shipperID,
		RecipientRole: entity.RoleShipper,
		Type:          entity.TypeOfferReceived,
		Channel:       entity.ChannelPush,
		Title:         "Bạn nhận được một báo giá mới",
		Body:          fmt.Sprintf("Một tài xế vừa báo giá %.0f đ cho đơn hàng của bạn.", payload.Price),
		RefType:       entity.RefTypeBid,
		RefID:         payload.BidID,
		Data: biz.MarshalData(map[string]string{
			"bid_id": payload.BidID,
			"ask_id": payload.AskID,
			"screen": "OfferList",
		}),
	}}, nil
}

func (c *MatchingConsumer) onOfferRejected(env *events.Envelope) ([]entity.CreateNotificationParam, error) {
	var payload events.OfferRejected
	if err := decodeData(env.Data, &payload); err != nil {
		return nil, err
	}

	driverID, err := uuid.Parse(payload.DriverID)
	if err != nil {
		return nil, fmt.Errorf("driver_id không hợp lệ: %w", err)
	}

	body := "Chủ hàng đã chọn tài xế khác cho đơn này."
	if payload.Reason != "" {
		body = payload.Reason
	}

	return []entity.CreateNotificationParam{{
		UserID:        driverID,
		RecipientRole: entity.RoleDriver,
		Type:          entity.TypeOfferRejected,
		Channel:       entity.ChannelInApp,
		Title:         "Báo giá của bạn không được chọn",
		Body:          body,
		RefType:       entity.RefTypeBid,
		RefID:         payload.BidID,
		Data: biz.MarshalData(map[string]string{
			"bid_id": payload.BidID,
			"ask_id": payload.AskID,
			"screen": "MyOffers",
		}),
	}}, nil
}

func (c *MatchingConsumer) onCargoSuggested(env *events.Envelope) ([]entity.CreateNotificationParam, error) {
	var payload events.CargoSuggested
	if err := decodeData(env.Data, &payload); err != nil {
		return nil, err
	}

	driverID, err := uuid.Parse(payload.DriverID)
	if err != nil {
		return nil, fmt.Errorf("driver_id không hợp lệ: %w", err)
	}

	return []entity.CreateNotificationParam{{
		UserID:        driverID,
		RecipientRole: entity.RoleDriver,
		Type:          entity.TypeCargoSuggested,
		Channel:       entity.ChannelPush,
		Title:         "Có đơn hàng cho chuyến của bạn",
		Body: fmt.Sprintf("Tìm thấy %d đơn hàng phù hợp với chuyến bạn vừa đăng. Xem và báo giá ngay.",
			payload.TotalFound),
		RefType: entity.RefTypeAsk,
		RefID:   payload.AskID,
		Data: biz.MarshalData(map[string]any{
			"ask_id":  payload.AskID,
			"bid_ids": payload.BidIDs,
			"screen":  "SuggestedCargo",
		}),
	}}, nil
}

func decodeData(data map[string]any, dest any) error {
	if len(data) == 0 {
		return errors.New("envelope không có trường data")
	}
	blob, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal lại data thất bại: %w", err)
	}
	if err := json.Unmarshal(blob, dest); err != nil {
		return fmt.Errorf("decode data thất bại: %w", err)
	}
	return nil
}