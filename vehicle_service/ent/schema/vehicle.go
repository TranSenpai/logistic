package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Vehicle holds the schema definition for the Vehicle entity.
type Vehicle struct {
	ent.Schema
}

// Fields of the Vehicle.
func (Vehicle) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Unique(),
		field.UUID("driver_id", uuid.UUID{}), // We store the User ID from user_service
		field.String("license_plate").Unique(),
		field.String("brand").Optional(),
		field.String("model").Optional(),
		field.Enum("vehicle_type").Values("truck", "van", "bike"),
		field.Float("capacity_weight_kg").Default(0.0),
		field.Float("capacity_volume_cbm").Default(0.0),
		field.Enum("status").Values("active", "maintenance").Default("active"),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the Vehicle.
func (Vehicle) Edges() []ent.Edge {
	return nil
}
