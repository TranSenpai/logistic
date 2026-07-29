package schema

import (
	"matching_service/ent/schema/mixin"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type Bids struct {
	ent.Schema
}

func (Bids) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(func() uuid.UUID { return uuid.Must(uuid.NewV7()) }),
		field.UUID("user_id", uuid.UUID{}),
		field.Float("volume_m3"),
		field.Float("weight_kg"),
		field.Float("max_price").Optional().Nillable(),
		field.String("zone_id"),
		field.Float("origin_lat"),
		field.Float("origin_lng"),
		field.Float("destination_lat"),
		field.Float("destination_lng"),
		field.Bytes("route_id"),
		field.Int8("status").Default(0),
	}
}

func (Bids) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("matches", Match.Type),
		edge.To("bids_requirements", Bids_Requirements.Type),
	}
}

func (Bids) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AuditMixin{},
		mixin.SoftDeleteMixin{},
	}
}
