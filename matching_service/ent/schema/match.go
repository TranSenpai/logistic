package schema

import (
	softdelete "matching_service/ent/softdelete"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type Match struct {
	ent.Schema
}

func (Match) Fields() []ent.Field {
	return []ent.Field{
		field.Int("bid_id"),
		field.Int("ask_id"),
		field.Float("agreed_price"),
		field.Int("status").Default(1), // 1: Proposed, 2: Accepted, 3: Rejected
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Match) Edges() []ent.Edge {
	return nil
}

func (Match) Mixin() []ent.Mixin {
	return []ent.Mixin{
		softdelete.SoftDeleteMixin{},
	}
}
