package schema

import (
	softdelete "matching_service/ent/softdelete"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/field"
)

type Bid struct {
	ent.Schema
}

func (Bid) Fields() []ent.Field {
	return []ent.Field{
		field.Int("user_id"),
		field.String("pickup_coordinates").
			SchemaType(map[string]string{dialect.Postgres: "geometry(Point, 4326)"}),
		field.String("delivery_coordinates").
			SchemaType(map[string]string{dialect.Postgres: "geometry(Point, 4326)"}),
		field.Float("volume_m3"),
		field.Float("weight_kg"),
		field.Time("pickup_time").Optional().Nillable(),
		field.Float("max_price").Optional().Nillable(),
		field.String("zone_id"),
		field.JSON("items", map[string]interface{}{}).Optional(),
		field.Int("status").Default(0), // 0: Pending, 1: Matched, 2: Expired, 3: Cancelled
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Bid) Edges() []ent.Edge {
	return nil
}

func (Bid) Mixin() []ent.Mixin {
	return []ent.Mixin{
		softdelete.SoftDeleteMixin{},
	}
}
