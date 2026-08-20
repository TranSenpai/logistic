package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Notification là MỘT thông báo gửi tới MỘT người nhận.
//
// Chủ ý thiết kế: sự kiện "đã tìm được xe" sinh ra HAI dòng — một cho chủ hàng
// ("đã tìm được xe cho đơn của bạn"), một cho tài xế ("bạn vừa nhận được đơn").
// Nếu gộp chung một dòng nhiều người nhận thì cờ is_read sẽ vô nghĩa, vì hai
// người đọc ở hai thời điểm khác nhau.
type Notification struct {
	ent.Schema
}

func (Notification) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Unique(),
		field.UUID("user_id", uuid.UUID{}),
		field.Enum("recipient_role").Values("driver", "shipper", "admin").Default("driver"),

		// type là mã nghiệp vụ, app dùng để chọn icon/màn hình đích.
		field.String("type"),
		field.Enum("channel").Values("in_app", "push", "email", "sms").Default("in_app"),

		field.String("title"),
		field.Text("body"),
		// data là JSON thô để app deep-link (vd: {"bid_id":"...","screen":"MatchDetail"}).
		field.Text("data").Optional(),

		// ref_type/ref_id trỏ ngược về đối tượng nghiệp vụ đã sinh ra thông báo.
		field.String("ref_type").Optional(),
		field.String("ref_id").Optional(),

		field.Bool("is_read").Default(false),
		field.Enum("status").Values("pending", "sent", "failed", "read").Default("pending"),
		field.String("error_message").Optional(),
		field.Time("read_at").Optional().Nillable(),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

func (Notification) Indexes() []ent.Index {
	return []ent.Index{
		// Truy vấn nóng nhất: "inbox của tôi, mới nhất trước".
		index.Fields("user_id", "created_at"),
		// Đếm số chưa đọc khi Redis miss.
		index.Fields("user_id", "is_read"),
		index.Fields("type"),
		index.Fields("status"),
	}
}
