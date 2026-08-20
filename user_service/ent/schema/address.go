package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Address là sổ địa chỉ của người dùng.
//
// Vì sao lưu sẵn latitude/longitude: khi shipper tạo đơn, matching_service cần
// toạ độ ngay để tính zone. Nếu chỉ lưu chữ thì mỗi lần đặt đơn lại phải gọi
// geocoding bên thứ ba — chậm và tốn tiền cho dữ liệu gần như không đổi.
type Address struct {
	ent.Schema
}

func (Address) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Unique(),
		field.UUID("user_id", uuid.UUID{}),
		field.String("label").Optional(),
		field.String("contact_name").Optional(),
		field.String("contact_phone").Optional(),
		field.String("line1"),
		field.String("ward").Optional(),
		field.String("district").Optional(),
		field.String("city").Optional(),
		field.Float("latitude").Default(0),
		field.Float("longitude").Default(0),
		field.Enum("address_type").Values("pickup", "delivery", "both").Default("both"),
		field.Bool("is_default").Default(false),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Address) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("addresses").Field("user_id").Unique().Required(),
	}
}

func (Address) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "address_type"),
		index.Fields("user_id", "is_default"),
	}
}
