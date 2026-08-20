package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// DriverProfile là phần hồ sơ chỉ tài xế mới có: bằng lái, CCCD, điểm đánh giá,
// trạng thái KYC.
type DriverProfile struct {
	ent.Schema
}

func (DriverProfile) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Unique(),
		// user_id khai báo tường minh để làm edge-field: entity có sẵn UserID,
		// repo lọc thẳng bằng WHERE user_id = ? thay vì phải JOIN sang bảng users.
		field.UUID("user_id", uuid.UUID{}),

		// Optional + Nillable: lúc đăng ký tài xế chưa có bằng lái, để NULL.
		// Nếu để "" thì unique index sẽ nổ ngay ở tài xế thứ hai — Postgres coi
		// nhiều NULL là khác nhau, nhưng nhiều chuỗi rỗng là trùng.
		field.String("license_number").Optional().Nillable().Unique(),
		field.String("id_card").Optional().Nillable().Unique(),

		field.Float("rating").Default(5.0),
		field.Int("total_trips").Default(0),
		field.Enum("kyc_status").Values("pending", "approved", "rejected").Default("pending"),
		field.String("kyc_note").Optional(),
		field.UUID("kyc_reviewed_by", uuid.UUID{}).Optional().Nillable(),
		field.Time("kyc_reviewed_at").Optional().Nillable(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (DriverProfile) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("driver_profile").Field("user_id").Unique().Required(),
	}
}

func (DriverProfile) Indexes() []ent.Index {
	return []ent.Index{
		// Hàng đợi duyệt KYC của admin quét theo cột này.
		index.Fields("kyc_status"),
		index.Fields("user_id").Unique(),
	}
}
