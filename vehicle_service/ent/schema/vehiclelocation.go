package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// VehicleLocation giữ vị trí GPS MỚI NHẤT của xe — mỗi xe đúng một dòng.
//
// Vì sao không lưu lịch sử ở đây: tài xế ping GPS vài giây một lần, nhân với
// hàng nghìn xe là hàng triệu dòng/ngày. Bảng này chỉ cần trả lời "xe đang ở
// đâu", nên ghi đè tại chỗ. Toạ độ nóng phục vụ tìm kiếm nằm trên Redis GEO;
// bảng này là bản lưu bền để khởi động lại vẫn dựng lại được index.
type VehicleLocation struct {
	ent.Schema
}

func (VehicleLocation) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Unique(),
		field.UUID("vehicle_id", uuid.UUID{}).Unique(),
		field.UUID("driver_id", uuid.UUID{}),
		field.Float("latitude"),
		field.Float("longitude"),
		field.Float("heading").Default(0),
		field.Float("speed_kph").Default(0),
		field.String("zone_id").Optional(),
		field.Time("recorded_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (VehicleLocation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("vehicle", Vehicle.Type).Ref("location").Field("vehicle_id").Unique().Required(),
	}
}

func (VehicleLocation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("zone_id"),
		index.Fields("driver_id"),
	}
}
