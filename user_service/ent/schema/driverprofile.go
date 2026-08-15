package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// DriverProfile holds the schema definition for the DriverProfile entity.
type DriverProfile struct {
	ent.Schema
}

// Fields of the DriverProfile.
func (DriverProfile) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Unique(),
		field.String("license_number").Unique(),
		field.String("id_card").Unique(),
		field.Float("rating").Default(5.0),
		field.Enum("kyc_status").Values("pending", "approved", "rejected").Default("pending"),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the DriverProfile.
func (DriverProfile) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("driver_profile").Unique().Required(),
	}
}
