package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// NotificationTemplate cho phép sửa câu chữ thông báo mà KHÔNG cần deploy lại.
//
// Placeholder viết theo dạng {{ten_bien}} và được điền lúc render. Nhờ vậy đội
// vận hành đổi lời nhắn hay thêm bản dịch tiếng Anh chỉ bằng một API admin.
type NotificationTemplate struct {
	ent.Schema
}

func (NotificationTemplate) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Unique(),
		field.String("code"), // MATCH_FOUND_SHIPPER, DRIVER_CANDIDATE...
		field.String("name"),
		field.Enum("channel").Values("in_app", "push", "email", "sms").Default("in_app"),
		field.String("locale").Default("vi"),
		field.String("title_template"),
		field.Text("body_template"),
		field.Bool("is_active").Default(true),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (NotificationTemplate) Indexes() []ent.Index {
	return []ent.Index{
		// Một mã template có nhiều biến thể theo kênh gửi và ngôn ngữ,
		// nhưng mỗi tổ hợp chỉ được tồn tại đúng một bản.
		index.Fields("code", "channel", "locale").Unique(),
		index.Fields("is_active"),
	}
}
