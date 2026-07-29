package schema

import (
	"matching_service/ent/schema/mixin"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type Match struct {
	ent.Schema
}

func (Match) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("bid_id", uuid.UUID{}).Comment("Logical FK point to Bids"),
		field.UUID("ask_id", uuid.UUID{}).Comment("Logical FK point to Asks"),
		field.Float("agreed_price"),
		field.Int("status").Default(1), // 1: Proposed, 2: Accepted, 3: Rejected
	}
}

func (Match) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("asks", Asks.Type).Ref("matches").Field("ask_id").Unique().Required(),
		edge.From("bids", Bids.Type).Ref("matches").Field("bid_id").Unique().Required(),
	}
}

func (Match) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AuditMixin{},
		mixin.SoftDeleteMixin{},
	}
}
