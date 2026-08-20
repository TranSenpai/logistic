package controller

import (
	"gateway_service/internal/middleware"
	"gateway_service/internal/response"

	"github.com/gin-gonic/gin"
	pb "github.com/logistic/api/logistic/notification_service/v1"
)

// NotificationController phơi phần ĐỌC của hộp thư thông báo ra HTTP.
//
// Không có endpoint "tạo thông báo" cho client — thông báo được sinh ra bởi
// consumer RabbitMQ khi matching_service phát sự kiện. Chỉ admin mới gửi thủ
// công được, qua nhóm /admin.
type NotificationController struct {
	notifClient pb.NotificationServiceClient
}

func NewNotificationController(notifClient pb.NotificationServiceClient) *NotificationController {
	return &NotificationController{notifClient: notifClient}
}

func notifUserID(ctx *gin.Context) string {
	if v := ctx.Param("user_id"); v != "" {
		return v
	}
	if v := ctx.Query("user_id"); v != "" {
		return v
	}
	return middleware.CurrentUserID(ctx)
}

// ===========================================================================
// CLIENT
// ===========================================================================

// ListNotifications godoc
// @Summary      Hộp thư thông báo
// @Tags         Notification
// @Produce      json
// @Param        user_id path string true "User ID"
// @Success      200 {object} response.Envelope
// @Router       /api/v1/users/{user_id}/notifications [get]
func (c *NotificationController) ListNotifications(ctx *gin.Context) {
	resp, err := c.notifClient.ListNotifications(ctx.Request.Context(), &pb.ListNotificationsRequest{
		UserId:     notifUserID(ctx),
		Type:       ctx.Query("type"),
		UnreadOnly: queryBool(ctx, "unread_only"),
		Page:       queryInt(ctx, "page"),
		PageSize:   queryInt(ctx, "page_size"),
	})
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.OK(ctx, gin.H{
		"notifications": resp.Notifications,
		"pagination":    resp.Pagination,
		"unread_count":  resp.UnreadCount,
	})
}

// GetNotification godoc
// @Summary      Chi tiết một thông báo
// @Tags         Notification
// @Produce      json
// @Param        id path string true "Notification ID"
// @Success      200 {object} response.Envelope
// @Router       /api/v1/notifications/{id} [get]
func (c *NotificationController) GetNotification(ctx *gin.Context) {
	resp, err := c.notifClient.GetNotification(ctx.Request.Context(), &pb.GetNotificationRequest{
		Id:     ctx.Param("id"),
		UserId: middleware.CurrentUserID(ctx),
	})
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, resp.Notification)
}

// MarkAsRead godoc
// @Summary      Đánh dấu đã đọc
// @Tags         Notification
// @Produce      json
// @Param        id path string true "Notification ID"
// @Success      200 {object} response.Envelope
// @Router       /api/v1/notifications/{id}/read [put]
func (c *NotificationController) MarkAsRead(ctx *gin.Context) {
	resp, err := c.notifClient.MarkAsRead(ctx.Request.Context(), &pb.MarkAsReadRequest{
		Id:     ctx.Param("id"),
		UserId: middleware.CurrentUserID(ctx),
	})
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OKMessage(ctx, gin.H{"unread_count": resp.UnreadCount}, resp.Message)
}

// MarkAllAsRead godoc
// @Summary      Đánh dấu tất cả đã đọc
// @Tags         Notification
// @Produce      json
// @Param        user_id path string true "User ID"
// @Success      200 {object} response.Envelope
// @Router       /api/v1/users/{user_id}/notifications/read-all [put]
func (c *NotificationController) MarkAllAsRead(ctx *gin.Context) {
	resp, err := c.notifClient.MarkAllAsRead(ctx.Request.Context(), &pb.MarkAllAsReadRequest{
		UserId: notifUserID(ctx),
	})
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OKMessage(ctx, gin.H{"marked_count": resp.MarkedCount}, resp.Message)
}

// DeleteNotification godoc
// @Summary      Xoá thông báo
// @Tags         Notification
// @Produce      json
// @Param        id path string true "Notification ID"
// @Success      200 {object} response.Envelope
// @Router       /api/v1/notifications/{id} [delete]
func (c *NotificationController) DeleteNotification(ctx *gin.Context) {
	resp, err := c.notifClient.DeleteNotification(ctx.Request.Context(), &pb.DeleteNotificationRequest{
		Id:     ctx.Param("id"),
		UserId: middleware.CurrentUserID(ctx),
	})
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OKMessage(ctx, nil, resp.Message)
}

// GetUnreadCount godoc
// @Summary      Số thông báo chưa đọc
// @Description  App gọi ở mọi màn hình để vẽ chấm đỏ; con số này được cache trên Redis.
// @Tags         Notification
// @Produce      json
// @Param        user_id path string true "User ID"
// @Success      200 {object} response.Envelope
// @Router       /api/v1/users/{user_id}/notifications/unread-count [get]
func (c *NotificationController) GetUnreadCount(ctx *gin.Context) {
	resp, err := c.notifClient.GetUnreadCount(ctx.Request.Context(), &pb.GetUnreadCountRequest{
		UserId: notifUserID(ctx),
	})
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, gin.H{"unread_count": resp.UnreadCount})
}

// GetPreferences godoc
// @Summary      Cài đặt nhận thông báo
// @Tags         Notification
// @Produce      json
// @Param        user_id path string true "User ID"
// @Success      200 {object} response.Envelope
// @Router       /api/v1/users/{user_id}/notification-preferences [get]
func (c *NotificationController) GetPreferences(ctx *gin.Context) {
	resp, err := c.notifClient.GetPreferences(ctx.Request.Context(), &pb.GetPreferencesRequest{
		UserId: notifUserID(ctx),
	})
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, resp.Preference)
}

type UpdatePreferencesReq struct {
	InAppEnabled       bool   `json:"in_app_enabled"`
	PushEnabled        bool   `json:"push_enabled"`
	EmailEnabled       bool   `json:"email_enabled"`
	SMSEnabled         bool   `json:"sms_enabled"`
	MatchEventsEnabled bool   `json:"match_events_enabled"`
	PromotionEnabled   bool   `json:"promotion_enabled"`
	QuietHoursStart    string `json:"quiet_hours_start"`
	QuietHoursEnd      string `json:"quiet_hours_end"`
}

// UpdatePreferences godoc
// @Summary      Cập nhật cài đặt nhận thông báo
// @Tags         Notification
// @Accept       json
// @Produce      json
// @Param        user_id path string true "User ID"
// @Success      200 {object} response.Envelope
// @Router       /api/v1/users/{user_id}/notification-preferences [put]
func (c *NotificationController) UpdatePreferences(ctx *gin.Context) {
	var req UpdatePreferencesReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "VALIDATION_FAILED", err.Error())
		return
	}

	resp, err := c.notifClient.UpdatePreferences(ctx.Request.Context(), &pb.UpdatePreferencesRequest{
		UserId:             notifUserID(ctx),
		InAppEnabled:       req.InAppEnabled,
		PushEnabled:        req.PushEnabled,
		EmailEnabled:       req.EmailEnabled,
		SmsEnabled:         req.SMSEnabled,
		MatchEventsEnabled: req.MatchEventsEnabled,
		PromotionEnabled:   req.PromotionEnabled,
		QuietHoursStart:    req.QuietHoursStart,
		QuietHoursEnd:      req.QuietHoursEnd,
	})
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OKMessage(ctx, resp.Preference, resp.Message)
}

// ===========================================================================
// ADMIN
// ===========================================================================

type AdminSendNotificationReq struct {
	UserIDs       []string `json:"user_ids"`
	BroadcastRole string   `json:"broadcast_role" binding:"omitempty,oneof=driver shipper admin"`
	Type          string   `json:"type"`
	Channel       string   `json:"channel" binding:"omitempty,oneof=in_app push email sms"`
	Title         string   `json:"title" binding:"required"`
	Body          string   `json:"body" binding:"required"`
	Data          string   `json:"data"`
}

// AdminSendNotification godoc
// @Summary      [Admin] Gửi thông báo thủ công
// @Tags         Admin-Notification
// @Accept       json
// @Produce      json
// @Success      200 {object} response.Envelope
// @Router       /api/v1/admin/notifications/send [post]
func (c *NotificationController) AdminSendNotification(ctx *gin.Context) {
	var req AdminSendNotificationReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "VALIDATION_FAILED", err.Error())
		return
	}

	resp, err := c.notifClient.AdminSendNotification(ctx.Request.Context(), &pb.AdminSendNotificationRequest{
		UserIds:       req.UserIDs,
		BroadcastRole: req.BroadcastRole,
		Type:          req.Type,
		Channel:       req.Channel,
		Title:         req.Title,
		Body:          req.Body,
		Data:          req.Data,
	})
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OKMessage(ctx, gin.H{"sent_count": resp.SentCount}, resp.Message)
}

// AdminListNotifications godoc
// @Summary      [Admin] Danh sách toàn bộ thông báo
// @Tags         Admin-Notification
// @Produce      json
// @Success      200 {object} response.Envelope
// @Router       /api/v1/admin/notifications [get]
func (c *NotificationController) AdminListNotifications(ctx *gin.Context) {
	resp, err := c.notifClient.AdminListNotifications(ctx.Request.Context(), &pb.AdminListNotificationsRequest{
		UserId:   ctx.Query("user_id"),
		Type:     ctx.Query("type"),
		Status:   ctx.Query("status"),
		Page:     queryInt(ctx, "page"),
		PageSize: queryInt(ctx, "page_size"),
	})
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, gin.H{"notifications": resp.Notifications, "pagination": resp.Pagination})
}

// AdminListTemplates godoc
// @Summary      [Admin] Danh sách template
// @Tags         Admin-Notification
// @Produce      json
// @Success      200 {object} response.Envelope
// @Router       /api/v1/admin/notification-templates [get]
func (c *NotificationController) AdminListTemplates(ctx *gin.Context) {
	resp, err := c.notifClient.AdminListTemplates(ctx.Request.Context(), &pb.AdminListTemplatesRequest{
		Channel: ctx.Query("channel"),
		Locale:  ctx.Query("locale"),
	})
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, gin.H{"templates": resp.Templates})
}

type CreateTemplateReq struct {
	Code          string `json:"code" binding:"required"`
	Name          string `json:"name" binding:"required"`
	Channel       string `json:"channel" binding:"omitempty,oneof=in_app push email sms"`
	Locale        string `json:"locale"`
	TitleTemplate string `json:"title_template" binding:"required"`
	BodyTemplate  string `json:"body_template" binding:"required"`
	IsActive      bool   `json:"is_active"`
}

// AdminCreateTemplate godoc
// @Summary      [Admin] Tạo template thông báo
// @Tags         Admin-Notification
// @Accept       json
// @Produce      json
// @Success      201 {object} response.Envelope
// @Router       /api/v1/admin/notification-templates [post]
func (c *NotificationController) AdminCreateTemplate(ctx *gin.Context) {
	var req CreateTemplateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "VALIDATION_FAILED", err.Error())
		return
	}

	resp, err := c.notifClient.AdminCreateTemplate(ctx.Request.Context(), &pb.AdminCreateTemplateRequest{
		Code:          req.Code,
		Name:          req.Name,
		Channel:       req.Channel,
		Locale:        req.Locale,
		TitleTemplate: req.TitleTemplate,
		BodyTemplate:  req.BodyTemplate,
		IsActive:      req.IsActive,
	})
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Created(ctx, resp.Template, resp.Message)
}

type UpdateTemplateReq struct {
	Name          string `json:"name"`
	TitleTemplate string `json:"title_template"`
	BodyTemplate  string `json:"body_template"`
	IsActive      bool   `json:"is_active"`
}

// AdminUpdateTemplate godoc
// @Summary      [Admin] Sửa template thông báo
// @Tags         Admin-Notification
// @Accept       json
// @Produce      json
// @Param        id path string true "Template ID"
// @Success      200 {object} response.Envelope
// @Router       /api/v1/admin/notification-templates/{id} [put]
func (c *NotificationController) AdminUpdateTemplate(ctx *gin.Context) {
	var req UpdateTemplateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "VALIDATION_FAILED", err.Error())
		return
	}

	resp, err := c.notifClient.AdminUpdateTemplate(ctx.Request.Context(), &pb.AdminUpdateTemplateRequest{
		Id:            ctx.Param("id"),
		Name:          req.Name,
		TitleTemplate: req.TitleTemplate,
		BodyTemplate:  req.BodyTemplate,
		IsActive:      req.IsActive,
	})
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OKMessage(ctx, resp.Template, resp.Message)
}

// AdminDeleteTemplate godoc
// @Summary      [Admin] Xoá template thông báo
// @Tags         Admin-Notification
// @Produce      json
// @Param        id path string true "Template ID"
// @Success      200 {object} response.Envelope
// @Router       /api/v1/admin/notification-templates/{id} [delete]
func (c *NotificationController) AdminDeleteTemplate(ctx *gin.Context) {
	resp, err := c.notifClient.AdminDeleteTemplate(ctx.Request.Context(), &pb.AdminDeleteTemplateRequest{
		Id: ctx.Param("id"),
	})
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OKMessage(ctx, nil, resp.Message)
}

// AdminGetNotificationStats godoc
// @Summary      [Admin] Thống kê thông báo
// @Tags         Admin-Notification
// @Produce      json
// @Success      200 {object} response.Envelope
// @Router       /api/v1/admin/notifications/stats [get]
func (c *NotificationController) AdminGetNotificationStats(ctx *gin.Context) {
	resp, err := c.notifClient.AdminGetNotificationStats(ctx.Request.Context(), &pb.AdminGetNotificationStatsRequest{})
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, resp)
}
