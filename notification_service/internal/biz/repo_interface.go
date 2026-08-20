package biz

import (
	"context"

	"notification_service/internal/entity"

	"github.com/google/uuid"
)

// NotificationRepo là cổng ra xuống Postgres + Redis của notification_service.
type NotificationRepo interface {
	// --- Notifications ---
	Create(ctx context.Context, param *entity.CreateNotificationParam) (*entity.Notification, error)
	// CreateBatch tạo nhiều thông báo trong MỘT transaction. Dùng cho luồng
	// fan-out: một sự kiện "có đơn phù hợp" đẻ ra N thông báo cho N tài xế.
	CreateBatch(ctx context.Context, params []entity.CreateNotificationParam) (int64, error)

	// CreateWithEventGuard tạo thông báo VÀ ghi dấu event_id trong cùng một
	// transaction. Trả về repo.ErrDuplicateEvent nếu event đã xử lý trước đó.
	// Đây là điểm chống trùng của toàn bộ luồng RabbitMQ.
	CreateWithEventGuard(ctx context.Context, eventID, routingKey, source string, params []entity.CreateNotificationParam) (int64, error)

	GetByID(ctx context.Context, id uuid.UUID) (*entity.Notification, error)
	List(ctx context.Context, param *entity.ListNotificationsParam) ([]entity.Notification, int64, error)
	AdminList(ctx context.Context, param *entity.AdminListNotificationsParam) ([]entity.Notification, int64, error)
	MarkAsRead(ctx context.Context, id uuid.UUID) (*entity.Notification, error)
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) (int64, error)
	Delete(ctx context.Context, id uuid.UUID) error
	CountUnread(ctx context.Context, userID uuid.UUID) (int64, error)
	CountAll(ctx context.Context) (int64, error)
	// CountUnreadAll đếm thông báo chưa đọc trên TOÀN hệ thống (cho màn thống kê
	// admin), khác với CountUnread vốn đếm theo từng người dùng.
	CountUnreadAll(ctx context.Context) (int64, error)
	CountByStatus(ctx context.Context, status string) (int64, error)
	CountSentToday(ctx context.Context) (int64, error)

	// --- Templates ---
	CreateTemplate(ctx context.Context, param *entity.CreateTemplateParam) (*entity.NotificationTemplate, error)
	GetTemplateByCode(ctx context.Context, code, channel, locale string) (*entity.NotificationTemplate, error)
	ListTemplates(ctx context.Context, param *entity.ListTemplatesParam) ([]entity.NotificationTemplate, error)
	UpdateTemplate(ctx context.Context, param *entity.UpdateTemplateParam) (*entity.NotificationTemplate, error)
	DeleteTemplate(ctx context.Context, id uuid.UUID) error
	CountTemplates(ctx context.Context) (int64, error)

	// --- Preferences ---
	// GetOrCreatePreference luôn trả về một bản ghi: người dùng chưa từng vào
	// phần cài đặt vẫn phải có mặc định để tầng biz quyết định gửi hay không.
	GetOrCreatePreference(ctx context.Context, userID uuid.UUID) (*entity.NotificationPreference, error)
	UpdatePreference(ctx context.Context, param *entity.UpdatePreferenceParam) (*entity.NotificationPreference, error)
}
