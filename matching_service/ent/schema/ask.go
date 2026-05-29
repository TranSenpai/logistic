package schema

import (
	softdelete "goBackend/matching_service/ent/softdelete"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/field"
)

type Ask struct {
	ent.Schema
}

func (Ask) Fields() []ent.Field {
	return []ent.Field{
		field.Int("driver_id"),
		field.String("current_coordinates").
			SchemaType(map[string]string{dialect.Postgres: "geometry(Point, 4326)"}).Optional(),
		field.Float("available_volume_m3"),
		field.Float("available_weight_kg"),
		field.Float("min_price").Optional().Nillable(),
		field.Int("status").Default(0), // 0: Pending, 1: Matched, 2: Offline
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Ask) Edges() []ent.Edge {
	return nil
}

func (Ask) Mixin() []ent.Mixin {
	return []ent.Mixin{
		softdelete.SoftDeleteMixin{},
	}
}
