package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Vehicle là phương tiện do tài xế đăng ký.
//
// Tách "status" (tài xế tự đổi: đang chạy / bảo dưỡng) khỏi "verification_status"
// (chỉ admin đổi được sau khi soi giấy tờ) là có chủ ý: tài xế không được phép
// tự tuyên bố xe mình đã hợp lệ.
type Vehicle struct {
	ent.Schema
}

func (Vehicle) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Unique(),
		field.UUID("driver_id", uuid.UUID{}), // User ID lấy từ user_service
		field.String("license_plate").Unique(),
		field.String("brand").Optional(),
		field.String("model").Optional(),
		field.Int("manufacture_year").Optional(),
		field.Enum("vehicle_type").Values("truck", "van", "bike", "container", "trailer"),
		field.Float("capacity_weight_kg").Default(0.0),
		field.Float("capacity_volume_cbm").Default(0.0),
		field.Enum("status").Values("active", "maintenance", "inactive").Default("active"),
		field.Enum("verification_status").Values("pending", "verified", "rejected").Default("pending"),
		field.String("verification_note").Optional(),
		field.UUID("verified_by", uuid.UUID{}).Optional().Nillable(),
		field.Time("verified_at").Optional().Nillable(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Vehicle) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("documents", VehicleDocument.Type),
		edge.To("location", VehicleLocation.Type).Unique(),
		edge.To("availability", DriverAvailability.Type).Unique(),
	}
}

func (Vehicle) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("driver_id", "status"),
		index.Fields("verification_status"),
		index.Fields("vehicle_type", "status"),
	}
}
