package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Unique(),
		field.String("phone").Unique(),
		field.String("email").Unique().Optional(),
		field.String("password_hash"),
		field.Enum("role").Values("driver", "shipper", "admin"),
		field.Enum("status").Values("active", "banned").Default("active"),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("driver_profile", DriverProfile.Type).Unique(),
		edge.To("shipper_profile", ShipperProfile.Type).Unique(),
	}
}
