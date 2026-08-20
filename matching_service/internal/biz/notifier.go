package biz

import (
	"context"

	"matching_service/internal/entity"
)

// Notifier là cổng ra để matching_service báo cho notification_service.
//
// Vì sao lại thêm một cổng nữa bên cạnh Kafka và NATS đang có:
//
//	Kafka -> nhật ký sự kiện lâu dài, dành cho phân tích và dựng lại trạng thái.
//	NATS  -> đẩy thời gian thực tới app đang mở (giá nhảy, danh sách cập nhật).
//	Notifier (RabbitMQ) -> thông báo BỀN cho người dùng: phải tới nơi kể cả khi
//	        app đang tắt, phải retry được, hỏng thì rơi vào dead-letter để người
//	        trực xem lại. NATS core không giữ message cho người đang offline,
//	        còn Kafka thì không có retry/DLQ theo từng message.
//
// Interface đặt ở tầng biz nên engine không hề biết bên dưới là RabbitMQ; khi
// test chỉ cần một implementation ghi vào slice là đủ.
type Notifier interface {
	// NotifyDriverCandidates: chủ hàng vừa đăng đơn, engine đã chấm ra danh sách
	// tài xế phù hợp -> báo cho TỪNG TÀI XẾ.
	NotifyDriverCandidates(ctx context.Context, bid *entity.Bid, asks []entity.Ask) error

	// NotifyMatchFound: đã chốt xe -> báo cho CẢ chủ hàng và tài xế.
	NotifyMatchFound(ctx context.Context, contract *entity.MatchContract, bid *entity.Bid, ask *entity.Ask) error

	// NotifyOfferReceived: tài xế vừa ra giá -> báo chủ hàng.
	NotifyOfferReceived(ctx context.Context, bid *entity.Bid, ask *entity.Ask, price float64) error

	// NotifyOfferRejected: chủ hàng từ chối giá -> báo tài xế.
	NotifyOfferRejected(ctx context.Context, bid *entity.Bid, ask *entity.Ask, reason string) error

	// NotifyCargoSuggested: tài xế đăng chuyến rỗng, engine tìm được đơn phù hợp
	// -> gợi ý cho chính tài xế đó.
	NotifyCargoSuggested(ctx context.Context, ask *entity.Ask, bids []entity.Bid) error
}

// NoopNotifier dùng khi RabbitMQ không dựng được. Nhờ nó, phần còn lại của
// engine không phải rải `if notifier != nil` ở khắp nơi — mất thông báo là điều
// đáng tiếc, nhưng việc ghép đơn vẫn phải chạy.
type NoopNotifier struct{}

func (NoopNotifier) NotifyDriverCandidates(context.Context, *entity.Bid, []entity.Ask) error {
	return nil
}

func (NoopNotifier) NotifyMatchFound(context.Context, *entity.MatchContract, *entity.Bid, *entity.Ask) error {
	return nil
}

func (NoopNotifier) NotifyOfferReceived(context.Context, *entity.Bid, *entity.Ask, float64) error {
	return nil
}

func (NoopNotifier) NotifyOfferRejected(context.Context, *entity.Bid, *entity.Ask, string) error {
	return nil
}

func (NoopNotifier) NotifyCargoSuggested(context.Context, *entity.Ask, []entity.Bid) error {
	return nil
}
