// Package mapper khai báo hợp đồng chuyển đổi dao <-> entity <-> dto của
// notification_service. Thân hàm do goverter sinh.
package mapper

import (
	"time"

	"github.com/google/uuid"
	pb "github.com/logistic/api/logistic/notification_service/v1"
	"google.golang.org/protobuf/types/known/timestamppb"

	"notification_service/ent"
	"notification_service/ent/notification"
	"notification_service/ent/notificationtemplate"
	"notification_service/internal/entity"
)

// goverter:converter
// goverter:matchIgnoreCase
// goverter:useZeroValueOnPointerInconsistency
// goverter:ignoreUnexported
// goverter:extend IdentityTime
// goverter:extend UUIDToString
// goverter:extend StringToUUID
// goverter:extend TimeToTimestamp
// goverter:extend TimestampToTime
// goverter:extend TimePtrToTime
// goverter:extend IntToInt32
// goverter:extend Int32ToInt
// goverter:extend EntRecipientRoleToString
// goverter:extend EntNotifChannelToString
// goverter:extend EntNotifStatusToString
// goverter:extend EntTemplateChannelToString
//
//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.4 gen ./
type AppMapper interface {
	// ==================== DAO -> ENTITY ====================

	EntNotificationToEntity(source *ent.Notification) entity.Notification
	EntNotificationListToEntityList(source []*ent.Notification) []entity.Notification

	EntTemplateToEntity(source *ent.NotificationTemplate) entity.NotificationTemplate
	EntTemplateListToEntityList(source []*ent.NotificationTemplate) []entity.NotificationTemplate

	EntPreferenceToEntity(source *ent.NotificationPreference) entity.NotificationPreference

	// ==================== ENTITY -> DTO ====================

	EntityNotificationToPb(source entity.Notification) *pb.Notification
	EntityNotificationListToPbList(source []entity.Notification) []*pb.Notification

	EntityTemplateToPb(source entity.NotificationTemplate) *pb.NotificationTemplate
	EntityTemplateListToPbList(source []entity.NotificationTemplate) []*pb.NotificationTemplate

	EntityPreferenceToPb(source entity.NotificationPreference) *pb.NotificationPreference

	EntityPaginationToPb(source entity.Pagination) *pb.Pagination

	// ==================== DTO -> ENTITY ====================

	PbListNotificationsToParam(req *pb.ListNotificationsRequest) (entity.ListNotificationsParam, error)

	PbAdminListNotificationsToParam(req *pb.AdminListNotificationsRequest) (entity.AdminListNotificationsParam, error)

	PbUpdatePreferencesToParam(req *pb.UpdatePreferencesRequest) (entity.UpdatePreferenceParam, error)

	PbListTemplatesToParam(req *pb.AdminListTemplatesRequest) entity.ListTemplatesParam

	PbCreateTemplateToParam(req *pb.AdminCreateTemplateRequest) entity.CreateTemplateParam

	// goverter:map Id ID
	PbUpdateTemplateToParam(req *pb.AdminUpdateTemplateRequest) (entity.UpdateTemplateParam, error)
}

// ===========================================================================
// HELPERS
// ===========================================================================

func IdentityTime(t time.Time) time.Time { return t }

func UUIDToString(u uuid.UUID) string {
	if u == uuid.Nil {
		return ""
	}
	return u.String()
}

func StringToUUID(s string) (uuid.UUID, error) {
	if s == "" {
		return uuid.Nil, nil
	}
	return uuid.Parse(s)
}

func TimeToTimestamp(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}

func TimestampToTime(ts *timestamppb.Timestamp) time.Time {
	if ts != nil && ts.IsValid() {
		return ts.AsTime()
	}
	return time.Time{}
}

// TimePtrToTime: read_at là cột nullable (chưa đọc thì NULL).
func TimePtrToTime(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}

func IntToInt32(i int) int32 { return int32(i) }
func Int32ToInt(i int32) int { return int(i) }

func EntRecipientRoleToString(r notification.RecipientRole) string { return string(r) }
func EntNotifChannelToString(c notification.Channel) string        { return string(c) }
func EntNotifStatusToString(s notification.Status) string          { return string(s) }
func EntTemplateChannelToString(c notificationtemplate.Channel) string {
	return string(c)
}
