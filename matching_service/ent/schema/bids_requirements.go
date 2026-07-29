package schema

import (
	"matching_service/ent/schema/mixin"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type Bids_Requirements struct {
	ent.Schema
}

func (Bids_Requirements) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(func() uuid.UUID { return uuid.Must(uuid.NewV7()) }),
		field.UUID("bids_id", uuid.UUID{}),
		field.UUID("requirements_id", uuid.UUID{}),
	}
}

func (Bids_Requirements) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("bids", Bids.Type).Ref("bids_requirements").Field("bids_id").Unique().Required(),
		edge.From("requirements", Requirements.Type).Ref("bids_requirements").Field("requirements_id").Unique().Required(),
	}
}

func (Bids_Requirements) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AuditMixin{},
		mixin.SoftDeleteMixin{},
	}
}
