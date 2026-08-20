// Package entity là tầng giữa dao <-> entity <-> dto của notification_service.
package entity

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// HẰNG SỐ
// ---------------------------------------------------------------------------

const (
	RoleDriver  = "driver"
	RoleShipper = "shipper"
	RoleAdmin   = "admin"
)

const (
	ChannelInApp = "in_app"
	ChannelPush  = "push"
	ChannelEmail = "email"
	ChannelSMS   = "sms"
)

const (
	StatusPending = "pending"
	StatusSent    = "sent"
	StatusFailed  = "failed"
	StatusRead    = "read"
)

// Mã loại thông báo. Đây là hợp đồng với app mobile: app dựa vào Type để chọn
// icon và màn hình mở ra khi người dùng bấm vào.
const (
	TypeDriverCandidate = "driver_candidate" // tài xế: có đơn hàng phù hợp gần bạn
	TypeMatchFound      = "match_found"      // cả hai phía: đã ghép được xe
	TypeOfferReceived   = "offer_received"   // chủ hàng: tài xế vừa báo giá
	TypeOfferRejected   = "offer_rejected"   // tài xế: chủ hàng từ chối giá
	TypeCargoSuggested  = "cargo_suggested"  // tài xế: gợi ý đơn cho chuyến rỗng
	TypeSystem          = "system"           // admin gửi thủ công
)

// Mã template mặc định, tương ứng với các loại ở trên.
const (
	TplDriverCandidate  = "DRIVER_CANDIDATE"
	TplMatchFoundShip   = "MATCH_FOUND_SHIPPER"
	TplMatchFoundDriver = "MATCH_FOUND_DRIVER"
	TplOfferReceived    = "OFFER_RECEIVED"
	TplOfferRejected    = "OFFER_REJECTED"
	TplCargoSuggested   = "CARGO_SUGGESTED"
)

const (
	RefTypeBid   = "bid"
	RefTypeAsk   = "ask"
	RefTypeMatch = "match"
)

func IsValidChannel(c string) bool {
	switch c {
	case ChannelInApp, ChannelPush, ChannelEmail, ChannelSMS:
		return true
	}
	return false
}

func IsValidRole(r string) bool {
	return r == RoleDriver || r == RoleShipper || r == RoleAdmin
}

func IsValidStatus(s string) bool {
	switch s {
	case StatusPending, StatusSent, StatusFailed, StatusRead:
		return true
	}
	return false
}

// ---------------------------------------------------------------------------
// DOMAIN ENTITIES
// ---------------------------------------------------------------------------

type Notification struct {
	ID            uuid.UUID `json:"id"`
	UserID        uuid.UUID `json:"user_id"`
	RecipientRole string    `json:"recipient_role"`
	Type          string    `json:"type"`
	Channel       string    `json:"channel"`
	Title         string    `json:"title"`
	Body          string    `json:"body"`
	Data          string    `json:"data"`
	RefType       string    `json:"ref_type"`
	RefID         string    `json:"ref_id"`
	IsRead        bool      `json:"is_read"`
	Status        string    `json:"status"`
	ErrorMessage  string    `json:"error_message"`
	ReadAt        time.Time `json:"read_at"`
	CreatedAt     time.Time `json:"created_at"`
}

type NotificationTemplate struct {
	ID            uuid.UUID `json:"id"`
	Code          string    `json:"code"`
	Name          string    `json:"name"`
	Channel       string    `json:"channel"`
	Locale        string    `json:"locale"`
	TitleTemplate string    `json:"title_template"`
	BodyTemplate  string    `json:"body_template"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Render thay các placeholder {{key}} bằng giá trị tương ứng.
//
// Cố tình KHÔNG dùng text/template: template do admin nhập qua API, mà
// text/template cho phép gọi hàm và truy cập field — mở đường cho việc một
// template sai (hoặc cố ý) làm panic tiến trình gửi thông báo. Thay chuỗi
// thuần thì trường hợp xấu nhất chỉ là câu chữ còn nguyên {{placeholder}}.
func (t *NotificationTemplate) Render(vars map[string]string) (string, string) {
	return replacePlaceholders(t.TitleTemplate, vars), replacePlaceholders(t.BodyTemplate, vars)
}

func replacePlaceholders(tpl string, vars map[string]string) string {
	if tpl == "" || len(vars) == 0 {
		return tpl
	}
	out := tpl
	for k, v := range vars {
		out = strings.ReplaceAll(out, "{{"+k+"}}", v)
	}
	return out
}

type NotificationPreference struct {
	ID                 uuid.UUID `json:"id"`
	UserID             uuid.UUID `json:"user_id"`
	InAppEnabled       bool      `json:"in_app_enabled"`
	PushEnabled        bool      `json:"push_enabled"`
	EmailEnabled       bool      `json:"email_enabled"`
	SMSEnabled         bool      `json:"sms_enabled"`
	MatchEventsEnabled bool      `json:"match_events_enabled"`
	PromotionEnabled   bool      `json:"promotion_enabled"`
	QuietHoursStart    string    `json:"quiet_hours_start"`
	QuietHoursEnd      string    `json:"quiet_hours_end"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// DefaultPreference dùng khi người dùng chưa từng đụng vào phần cài đặt.
// Mặc định bật in-app và push cho sự kiện ghép đơn — đó là lý do họ cài app.
func DefaultPreference(userID uuid.UUID) NotificationPreference {
	return NotificationPreference{
		UserID:             userID,
		InAppEnabled:       true,
		PushEnabled:        true,
		EmailEnabled:       false,
		SMSEnabled:         false,
		MatchEventsEnabled: true,
		PromotionEnabled:   true,
	}
}

// AllowsChannel trả lời "có được gửi qua kênh này không".
func (p *NotificationPreference) AllowsChannel(channel string) bool {
	switch channel {
	case ChannelInApp:
		return p.InAppEnabled
	case ChannelPush:
		return p.PushEnabled
	case ChannelEmail:
		return p.EmailEnabled
	case ChannelSMS:
		return p.SMSEnabled
	default:
		return false
	}
}

// AllowsType lọc theo NHÓM sự kiện, độc lập với kênh gửi. Một người có thể vẫn
// muốn nhận thông báo ghép đơn nhưng tắt hết khuyến mãi.
func (p *NotificationPreference) AllowsType(notiType string) bool {
	switch notiType {
	case TypeDriverCandidate, TypeMatchFound, TypeOfferReceived, TypeOfferRejected, TypeCargoSuggested:
		return p.MatchEventsEnabled
	case TypeSystem:
		return true // thông báo hệ thống luôn được gửi
	default:
		return p.PromotionEnabled
	}
}

// IsQuietHour cho biết thời điểm t có rơi vào khung giờ yên lặng không.
//
// Xử lý được cả khung qua đêm (22:00 -> 07:00): khi start > end thì khoảng
// thời gian bị "gấp" qua nửa đêm, nên điều kiện là HOẶC chứ không phải VÀ.
func (p *NotificationPreference) IsQuietHour(t time.Time) bool {
	if p.QuietHoursStart == "" || p.QuietHoursEnd == "" {
		return false
	}

	start, ok1 := parseHourMinute(p.QuietHoursStart)
	end, ok2 := parseHourMinute(p.QuietHoursEnd)
	if !ok1 || !ok2 {
		return false
	}

	now := t.Hour()*60 + t.Minute()
	if start <= end {
		return now >= start && now < end
	}
	return now >= start || now < end
}

func parseHourMinute(s string) (int, bool) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return 0, false
	}
	h, err1 := atoi(parts[0])
	m, err2 := atoi(parts[1])
	if !err1 || !err2 || h < 0 || h > 23 || m < 0 || m > 59 {
		return 0, false
	}
	return h*60 + m, true
}

func atoi(s string) (int, bool) {
	n := 0
	if s == "" {
		return 0, false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, false
		}
		n = n*10 + int(r-'0')
	}
	return n, true
}

// ---------------------------------------------------------------------------
// PHÂN TRANG
// ---------------------------------------------------------------------------

type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}

const (
	defaultPage     = 1
	defaultPageSize = 20
	maxPageSize     = 100
)

func NormalizePaging(page, pageSize int) (int, int, int) {
	if page < 1 {
		page = defaultPage
	}
	if pageSize < 1 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return page, pageSize, (page - 1) * pageSize
}

func BuildPagination(page, pageSize int, total int64) Pagination {
	totalPages := 0
	if pageSize > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}
	return Pagination{Page: page, PageSize: pageSize, TotalItems: total, TotalPages: totalPages}
}

// ---------------------------------------------------------------------------
// PARAMS & RESULTS
// ---------------------------------------------------------------------------

// CreateNotificationParam là đầu vào của luồng tạo thông báo. Luồng này KHÔNG
// có API public — nó chỉ được gọi từ consumer RabbitMQ và từ API admin.
type CreateNotificationParam struct {
	UserID        uuid.UUID
	RecipientRole string
	Type          string
	Channel       string
	Title         string
	Body          string
	Data          string
	RefType       string
	RefID         string
}

type ListNotificationsParam struct {
	UserID     uuid.UUID
	Type       string
	UnreadOnly bool
	Page       int
	PageSize   int
}

type ListNotificationsResult struct {
	Notifications []Notification
	Pagination    Pagination
	UnreadCount   int64
}

type AdminListNotificationsParam struct {
	UserID   uuid.UUID
	Type     string
	Status   string
	Page     int
	PageSize int
}

type UpdatePreferenceParam struct {
	UserID             uuid.UUID
	InAppEnabled       bool
	PushEnabled        bool
	EmailEnabled       bool
	SMSEnabled         bool
	MatchEventsEnabled bool
	PromotionEnabled   bool
	QuietHoursStart    string
	QuietHoursEnd      string
}

type ListTemplatesParam struct {
	Channel string
	Locale  string
}

type CreateTemplateParam struct {
	Code          string
	Name          string
	Channel       string
	Locale        string
	TitleTemplate string
	BodyTemplate  string
	IsActive      bool
}

type UpdateTemplateParam struct {
	ID            uuid.UUID
	Name          string
	TitleTemplate string
	BodyTemplate  string
	IsActive      bool
}

type SendNotificationParam struct {
	UserIDs       []uuid.UUID
	BroadcastRole string
	Type          string
	Channel       string
	Title         string
	Body          string
	Data          string
}

type SendNotificationResult struct {
	SentCount int64
	Message   string
}

type NotificationStats struct {
	TotalNotifications  int64
	UnreadNotifications int64
	SentToday           int64
	FailedNotifications int64
	TotalTemplates      int64
}
