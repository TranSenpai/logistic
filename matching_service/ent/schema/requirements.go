package schema

import (
	"github.com/logistic/pkg/ent/mixin"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type Requirements struct {
	ent.Schema
}

func (Requirements) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(func() uuid.UUID { return uuid.Must(uuid.NewV7()) }),
		field.String("name"),
		field.Int8("status"),
	}
}

func (Requirements) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("bids_requirements", Bids_Requirements.Type),
	}
}

func (Requirements) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AuditMixin{},
		mixin.SoftDeleteMixin{},
	}
}
