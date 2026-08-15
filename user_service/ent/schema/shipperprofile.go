package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// ShipperProfile holds the schema definition for the ShipperProfile entity.
type ShipperProfile struct {
	ent.Schema
}

// Fields of the ShipperProfile.
func (ShipperProfile) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Unique(),
		field.String("company_name").Optional(),
		field.String("tax_code").Optional(),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the ShipperProfile.
func (ShipperProfile) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("shipper_profile").Unique().Required(),
	}
}
