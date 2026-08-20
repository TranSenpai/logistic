package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// NotificationPreference là lựa chọn nhận thông báo của từng người dùng.
//
// quiet_hours_* để tài xế không bị đánh thức lúc 2 giờ sáng vì một đơn hàng.
// Thông báo trong giờ yên lặng vẫn được lưu vào inbox, chỉ không đẩy push.
type NotificationPreference struct {
	ent.Schema
}

func (NotificationPreference) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Unique(),
		field.UUID("user_id", uuid.UUID{}).Unique(),
		field.Bool("in_app_enabled").Default(true),
		field.Bool("push_enabled").Default(true),
		field.Bool("email_enabled").Default(false),
		field.Bool("sms_enabled").Default(false),
		field.Bool("match_events_enabled").Default(true),
		field.Bool("promotion_enabled").Default(true),
		field.String("quiet_hours_start").Optional(), // "22:00"
		field.String("quiet_hours_end").Optional(),   // "07:00"
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (NotificationPreference) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id").Unique(),
	}
}
