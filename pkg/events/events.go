// Package events là HỢP ĐỒNG chung giữa producer và consumer trên RabbitMQ.
//
// Cả matching_service (bên phát) và notification_service (bên nhận) đều import
// package này, nên không thể xảy ra cảnh một bên đổi tên field JSON còn bên kia
// vẫn parse theo tên cũ — trình biên dịch sẽ chặn trước.
package events

import "time"

// Tên exchange dùng chung. Kiểu topic để consumer tự chọn routing key cần nghe.
const ExchangeLogistic = "logistic.events"

// Routing key. Quy ước: <miền>.<đối tượng>.<việc đã xảy ra>, luôn ở THÌ QUÁ KHỨ
// vì event mô tả sự việc đã rồi, không phải câu lệnh.
const (
	// Shipper vừa đăng đơn, matching đã tính ra danh sách tài xế tiềm năng.
	// -> notification_service báo cho TỪNG TÀI XẾ "có đơn hàng phù hợp".
	RoutingKeyDriverCandidatesFound = "matching.driver.candidates_found"

	// Đã chốt được xe cho đơn hàng.
	// -> báo cho shipper "đã tìm được xe", đồng thời báo tài xế "bạn đã nhận đơn".
	RoutingKeyMatchFound = "matching.match.found"

	// Tài xế vừa ra giá cho đơn hàng -> báo shipper.
	RoutingKeyOfferReceived = "matching.offer.received"

	// Shipper từ chối báo giá -> báo tài xế.
	RoutingKeyOfferRejected = "matching.offer.rejected"

	// Tài xế vừa đăng chuyến rỗng và hệ thống tìm ra đơn phù hợp -> gợi ý cho tài xế.
	RoutingKeyCargoSuggested = "matching.cargo.suggested"
)

// Kênh gửi thông báo tới người dùng.
const (
	ChannelInApp = "in_app"
	ChannelPush  = "push"
	ChannelEmail = "email"
	ChannelSMS   = "sms"
)

// Envelope là lớp vỏ chung của MỌI message. Consumer luôn unmarshal ra Envelope
// trước, nhìn EventType rồi mới decode Data theo đúng kiểu.
type Envelope struct {
	EventID    string         `json:"event_id"`
	EventType  string         `json:"event_type"`
	OccurredAt time.Time      `json:"occurred_at"`
	TraceID    string         `json:"trace_id,omitempty"`
	Source     string         `json:"source"`
	Data       map[string]any `json:"data"`
}

// DriverCandidate là một tài xế được engine chấm là phù hợp với đơn hàng.
type DriverCandidate struct {
	DriverID   string  `json:"driver_id"`
	AskID      string  `json:"ask_id"`
	VehicleID  string  `json:"vehicle_id"`
	Phone      string  `json:"phone,omitempty"`
	Email      string  `json:"email,omitempty"`
	DistanceKm float64 `json:"distance_km"`
	Score      float64 `json:"score"`
}

// DriverCandidatesFound: shipper đăng đơn -> đây là danh sách tài xế cần được báo.
type DriverCandidatesFound struct {
	BidID          string            `json:"bid_id"`
	ShipperID      string            `json:"shipper_id"`
	OriginZoneID   string            `json:"origin_zone_id"`
	OriginLat      float64           `json:"origin_lat"`
	OriginLng      float64           `json:"origin_lng"`
	DestinationLat float64           `json:"destination_lat"`
	DestinationLng float64           `json:"destination_lng"`
	WeightKg       float64           `json:"weight_kg"`
	VolumeM3       float64           `json:"volume_m3"`
	MaxPrice       float64           `json:"max_price"`
	Candidates     []DriverCandidate `json:"candidates"`
}

// MatchFound: đã chốt xe. Cả hai phía đều cần được thông báo.
type MatchFound struct {
	ContractID       string  `json:"contract_id"`
	BidID            string  `json:"bid_id"`
	AskID            string  `json:"ask_id"`
	ShipperID        string  `json:"shipper_id"`
	DriverID         string  `json:"driver_id"`
	VehicleID        string  `json:"vehicle_id"`
	ConsensusPrice   float64 `json:"consensus_price"`
	ConsensusDeposit float64 `json:"consensus_deposit"`
	ShipperPhone     string  `json:"shipper_phone,omitempty"`
	ShipperEmail     string  `json:"shipper_email,omitempty"`
	DriverPhone      string  `json:"driver_phone,omitempty"`
	DriverEmail      string  `json:"driver_email,omitempty"`
}

// OfferReceived: tài xế ra giá cho đơn của shipper.
type OfferReceived struct {
	BidID     string  `json:"bid_id"`
	AskID     string  `json:"ask_id"`
	ShipperID string  `json:"shipper_id"`
	DriverID  string  `json:"driver_id"`
	Price     float64 `json:"price"`
}

// OfferRejected: shipper từ chối giá của tài xế.
type OfferRejected struct {
	BidID    string `json:"bid_id"`
	AskID    string `json:"ask_id"`
	DriverID string `json:"driver_id"`
	Reason   string `json:"reason,omitempty"`
}

// CargoSuggested: tài xế đăng chuyến rỗng -> gợi ý các đơn hàng phù hợp.
type CargoSuggested struct {
	AskID      string   `json:"ask_id"`
	DriverID   string   `json:"driver_id"`
	VehicleID  string   `json:"vehicle_id"`
	BidIDs     []string `json:"bid_ids"`
	TotalFound int      `json:"total_found"`
}
