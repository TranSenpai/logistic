// Package biz chứa luật nghiệp vụ của notification_service.
//
// Điểm khác biệt so với hai service kia: tầng này có HAI nguồn đầu vào.
//
//	gRPC     -> người dùng đọc inbox, admin quản lý template (luồng đọc).
//	RabbitMQ -> matching_service phát sự kiện, ta sinh thông báo (luồng ghi).
//
// Cả hai đều đi qua cùng một NotificationEngine, nên quy tắc "tôn trọng cài đặt
// nhận thông báo của người dùng" chỉ tồn tại đúng một chỗ.
package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	cerr "notification_service/internal/common/errors"
	"notification_service/internal/entity"

	"github.com/google/uuid"
)

type NotificationEngine interface {
	// --- Luồng đọc (gRPC client) ---
	List(ctx context.Context, param *entity.ListNotificationsParam) (*entity.ListNotificationsResult, error)
	Get(ctx context.Context, id, userID uuid.UUID) (*entity.Notification, error)
	MarkAsRead(ctx context.Context, id, userID uuid.UUID) (int64, error)
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) (int64, error)
	Delete(ctx context.Context, id, userID uuid.UUID) error
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error)
	GetPreference(ctx context.Context, userID uuid.UUID) (*entity.NotificationPreference, error)
	UpdatePreference(ctx context.Context, param *entity.UpdatePreferenceParam) (*entity.NotificationPreference, error)

	// --- Luồng ghi (consumer + admin) ---
	// DispatchEvent là điểm vào của consumer RabbitMQ: nó lọc theo cài đặt của
	// từng người nhận rồi ghi tất cả trong một transaction có chống trùng.
	DispatchEvent(ctx context.Context, eventID, routingKey, source string, params []entity.CreateNotificationParam) (int64, error)
	AdminSend(ctx context.Context, param *entity.SendNotificationParam) (*entity.SendNotificationResult, error)

	// --- Admin ---
	AdminList(ctx context.Context, param *entity.AdminListNotificationsParam) (*entity.ListNotificationsResult, error)
	AdminListTemplates(ctx context.Context, param *entity.ListTemplatesParam) ([]entity.NotificationTemplate, error)
	AdminCreateTemplate(ctx context.Context, param *entity.CreateTemplateParam) (*entity.NotificationTemplate, error)
	AdminUpdateTemplate(ctx context.Context, param *entity.UpdateTemplateParam) (*entity.NotificationTemplate, error)
	AdminDeleteTemplate(ctx context.Context, id uuid.UUID) error
	AdminGetStats(ctx context.Context) (*entity.NotificationStats, error)

	// RenderFromTemplate lấy template theo mã và điền biến. Không có template
	// trong DB thì trả về (false, ...) để caller dùng câu chữ mặc định trong code.
	RenderFromTemplate(ctx context.Context, code, channel, locale string, vars map[string]string) (string, string, bool)
}

type notificationEngineImpl struct {
	repo NotificationRepo
}

func NewNotificationEngine(repo NotificationRepo) NotificationEngine {
	return &notificationEngineImpl{repo: repo}
}

// ---------------------------------------------------------------------------
// LUỒNG ĐỌC
// ---------------------------------------------------------------------------

func (e *notificationEngineImpl) List(ctx context.Context, param *entity.ListNotificationsParam) (*entity.ListNotificationsResult, error) {
	if param.UserID == uuid.Nil {
		return nil, cerr.ErrInvalidUserID
	}

	page, pageSize, _ := entity.NormalizePaging(param.Page, param.PageSize)
	list, total, err := e.repo.List(ctx, param)
	if err != nil {
		return nil, err
	}

	unread, err := e.repo.CountUnread(ctx, param.UserID)
	if err != nil {
		return nil, err
	}

	return &entity.ListNotificationsResult{
		Notifications: list,
		Pagination:    entity.BuildPagination(page, pageSize, total),
		UnreadCount:   unread,
	}, nil
}

func (e *notificationEngineImpl) Get(ctx context.Context, id, userID uuid.UUID) (*entity.Notification, error) {
	if id == uuid.Nil {
		return nil, cerr.ErrInvalidNotifID
	}

	n, err := e.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	// Inbox là dữ liệu riêng tư: đoán đúng id vẫn không được đọc của người khác.
	if userID != uuid.Nil && n.UserID != userID {
		return nil, cerr.ErrNotificationNotOwned
	}
	return n, nil
}

func (e *notificationEngineImpl) MarkAsRead(ctx context.Context, id, userID uuid.UUID) (int64, error) {
	n, err := e.Get(ctx, id, userID)
	if err != nil {
		return 0, err
	}

	if !n.IsRead {
		if _, err := e.repo.MarkAsRead(ctx, id); err != nil {
			return 0, err
		}
	}
	return e.repo.CountUnread(ctx, n.UserID)
}

func (e *notificationEngineImpl) MarkAllAsRead(ctx context.Context, userID uuid.UUID) (int64, error) {
	if userID == uuid.Nil {
		return 0, cerr.ErrInvalidUserID
	}
	return e.repo.MarkAllAsRead(ctx, userID)
}

func (e *notificationEngineImpl) Delete(ctx context.Context, id, userID uuid.UUID) error {
	if _, err := e.Get(ctx, id, userID); err != nil {
		return err
	}
	return e.repo.Delete(ctx, id)
}

func (e *notificationEngineImpl) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	if userID == uuid.Nil {
		return 0, cerr.ErrInvalidUserID
	}
	return e.repo.CountUnread(ctx, userID)
}

func (e *notificationEngineImpl) GetPreference(ctx context.Context, userID uuid.UUID) (*entity.NotificationPreference, error) {
	if userID == uuid.Nil {
		return nil, cerr.ErrInvalidUserID
	}
	return e.repo.GetOrCreatePreference(ctx, userID)
}

func (e *notificationEngineImpl) UpdatePreference(ctx context.Context, param *entity.UpdatePreferenceParam) (*entity.NotificationPreference, error) {
	if param.UserID == uuid.Nil {
		return nil, cerr.ErrInvalidUserID
	}
	return e.repo.UpdatePreference(ctx, param)
}

// ---------------------------------------------------------------------------
// LUỒNG GHI
// ---------------------------------------------------------------------------

// DispatchEvent nhận danh sách thông báo cần tạo, lọc theo cài đặt của từng
// người nhận, rồi ghi tất cả trong một transaction có chống trùng theo eventID.
//
// Lọc TRƯỚC khi ghi chứ không phải trước khi gửi: người đã tắt nhận thông báo
// ghép đơn thì không nên thấy chúng nằm trong inbox chờ sẵn.
func (e *notificationEngineImpl) DispatchEvent(
	ctx context.Context,
	eventID, routingKey, source string,
	params []entity.CreateNotificationParam,
) (int64, error) {
	if eventID == "" {
		return 0, cerr.ErrInvalidNotifID.WithMessage("event_id là bắt buộc để chống xử lý trùng")
	}
	if len(params) == 0 {
		return 0, nil
	}

	allowed := make([]entity.CreateNotificationParam, 0, len(params))
	now := time.Now()

	for i := range params {
		p := params[i]
		if p.UserID == uuid.Nil {
			continue
		}
		if p.Channel == "" {
			p.Channel = entity.ChannelInApp
		}

		pref, err := e.repo.GetOrCreatePreference(ctx, p.UserID)
		if err != nil {
			// Không đọc được cài đặt thì vẫn gửi: bỏ sót một thông báo ghép đơn
			// tốn kém hơn nhiều so với gửi thừa một thông báo.
			log.Printf("[biz] không đọc được cài đặt của %s (%v) — vẫn gửi theo mặc định", p.UserID, err)
			allowed = append(allowed, p)
			continue
		}

		if !pref.AllowsType(p.Type) || !pref.AllowsChannel(p.Channel) {
			continue
		}

		// Giờ yên lặng chỉ chặn PUSH. Thông báo vẫn vào inbox để sáng hôm sau
		// người dùng mở app là thấy.
		if p.Channel == entity.ChannelPush && pref.IsQuietHour(now) {
			p.Channel = entity.ChannelInApp
		}

		allowed = append(allowed, p)
	}

	if len(allowed) == 0 {
		// Vẫn phải ghi dấu event: nếu không, message sẽ được xử lý lại mãi mỗi
		// lần broker giao lại, dù kết quả luôn là "không ai muốn nhận".
		return e.repo.CreateWithEventGuard(ctx, eventID, routingKey, source, nil)
	}

	return e.repo.CreateWithEventGuard(ctx, eventID, routingKey, source, allowed)
}

func (e *notificationEngineImpl) AdminSend(ctx context.Context, param *entity.SendNotificationParam) (*entity.SendNotificationResult, error) {
	if param.Title == "" {
		return nil, cerr.ErrTitleRequired
	}
	if param.Body == "" {
		return nil, cerr.ErrBodyRequired
	}
	if len(param.UserIDs) == 0 && param.BroadcastRole == "" {
		return nil, cerr.ErrNoRecipient
	}
	if param.BroadcastRole != "" && !entity.IsValidRole(param.BroadcastRole) {
		return nil, cerr.ErrInvalidRole.WithDetail("broadcast_role", param.BroadcastRole)
	}

	channel := param.Channel
	if channel == "" {
		channel = entity.ChannelInApp
	}
	if !entity.IsValidChannel(channel) {
		return nil, cerr.ErrInvalidChannel.WithDetail("channel", channel)
	}

	notiType := param.Type
	if notiType == "" {
		notiType = entity.TypeSystem
	}

	role := param.BroadcastRole
	if role == "" {
		role = entity.RoleDriver
	}

	params := make([]entity.CreateNotificationParam, 0, len(param.UserIDs))
	for _, uid := range param.UserIDs {
		if uid == uuid.Nil {
			continue
		}
		params = append(params, entity.CreateNotificationParam{
			UserID:        uid,
			RecipientRole: role,
			Type:          notiType,
			Channel:       channel,
			Title:         param.Title,
			Body:          param.Body,
			Data:          param.Data,
		})
	}

	if len(params) == 0 {
		// Broadcast theo vai trò cần danh sách người dùng từ user_service.
		// Chưa nối luồng đó nên từ chối rõ ràng thay vì âm thầm không gửi gì.
		return nil, cerr.ErrNoRecipient.WithMessage(
			"broadcast theo vai trò chưa được hỗ trợ — hãy truyền danh sách user_ids cụ thể")
	}

	sent, err := e.repo.CreateBatch(ctx, params)
	if err != nil {
		return nil, err
	}

	return &entity.SendNotificationResult{
		SentCount: sent,
		Message:   fmt.Sprintf("Đã gửi %d thông báo", sent),
	}, nil
}

// ---------------------------------------------------------------------------
// ADMIN
// ---------------------------------------------------------------------------

func (e *notificationEngineImpl) AdminList(ctx context.Context, param *entity.AdminListNotificationsParam) (*entity.ListNotificationsResult, error) {
	if param.Status != "" && !entity.IsValidStatus(param.Status) {
		return nil, cerr.ErrInvalidStatus.WithDetail("status", param.Status)
	}

	page, pageSize, _ := entity.NormalizePaging(param.Page, param.PageSize)
	list, total, err := e.repo.AdminList(ctx, param)
	if err != nil {
		return nil, err
	}

	return &entity.ListNotificationsResult{
		Notifications: list,
		Pagination:    entity.BuildPagination(page, pageSize, total),
	}, nil
}

func (e *notificationEngineImpl) AdminListTemplates(ctx context.Context, param *entity.ListTemplatesParam) ([]entity.NotificationTemplate, error) {
	if param.Channel != "" && !entity.IsValidChannel(param.Channel) {
		return nil, cerr.ErrInvalidChannel.WithDetail("channel", param.Channel)
	}
	return e.repo.ListTemplates(ctx, param)
}

func (e *notificationEngineImpl) AdminCreateTemplate(ctx context.Context, param *entity.CreateTemplateParam) (*entity.NotificationTemplate, error) {
	if param.Code == "" {
		return nil, cerr.ErrCodeRequired
	}
	if param.TitleTemplate == "" {
		return nil, cerr.ErrTitleRequired
	}
	if param.BodyTemplate == "" {
		return nil, cerr.ErrBodyRequired
	}
	if param.Channel != "" && !entity.IsValidChannel(param.Channel) {
		return nil, cerr.ErrInvalidChannel.WithDetail("channel", param.Channel)
	}
	return e.repo.CreateTemplate(ctx, param)
}

func (e *notificationEngineImpl) AdminUpdateTemplate(ctx context.Context, param *entity.UpdateTemplateParam) (*entity.NotificationTemplate, error) {
	if param.ID == uuid.Nil {
		return nil, cerr.ErrInvalidTemplateID
	}
	return e.repo.UpdateTemplate(ctx, param)
}

func (e *notificationEngineImpl) AdminDeleteTemplate(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return cerr.ErrInvalidTemplateID
	}
	return e.repo.DeleteTemplate(ctx, id)
}

func (e *notificationEngineImpl) AdminGetStats(ctx context.Context) (*entity.NotificationStats, error) {
	total, err := e.repo.CountAll(ctx)
	if err != nil {
		return nil, err
	}
	unread, err := e.repo.CountUnreadAll(ctx)
	if err != nil {
		return nil, err
	}
	sentToday, err := e.repo.CountSentToday(ctx)
	if err != nil {
		return nil, err
	}
	failed, err := e.repo.CountByStatus(ctx, entity.StatusFailed)
	if err != nil {
		return nil, err
	}
	templates, err := e.repo.CountTemplates(ctx)
	if err != nil {
		return nil, err
	}

	return &entity.NotificationStats{
		TotalNotifications:  total,
		UnreadNotifications: unread,
		SentToday:           sentToday,
		FailedNotifications: failed,
		TotalTemplates:      templates,
	}, nil
}

func (e *notificationEngineImpl) RenderFromTemplate(ctx context.Context, code, channel, locale string, vars map[string]string) (string, string, bool) {
	if locale == "" {
		locale = "vi"
	}
	if channel == "" {
		channel = entity.ChannelInApp
	}

	tpl, err := e.repo.GetTemplateByCode(ctx, code, channel, locale)
	if err != nil {
		return "", "", false
	}

	title, body := tpl.Render(vars)
	return title, body, true
}

// MarshalData gói payload thành JSON để nhét vào cột data.
// Lỗi marshal không được làm hỏng cả thông báo — mất deep-link còn hơn mất tin.
func MarshalData(v any) string {
	blob, err := json.Marshal(v)
	if err != nil {
		log.Printf("[biz] marshal notification data thất bại: %v", err)
		return ""
	}
	return string(blob)
}
