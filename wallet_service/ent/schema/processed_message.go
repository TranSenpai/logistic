package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/logistic/pkg/ent/mixin"
)

// ProcessedMessage holds the schema definition for the ProcessedMessage entity.
// Sử dụng làm Inbox Pattern / Idempotent Consumer để chống xử lý trùng lặp Kafka message.
type ProcessedMessage struct {
	ent.Schema
}

// Fields of the ProcessedMessage.
func (ProcessedMessage) Fields() []ent.Field {
	return []ent.Field{
		// ID chính là Message ID của Kafka (vd: UUID hoặc Partition-Offset)
		// Đảm bảo Unique() là cốt lõi của Idempotent.
		field.String("id").Unique().Immutable(),
	}
}

// Edges of the ProcessedMessage.
func (ProcessedMessage) Edges() []ent.Edge {
	return nil
}

func (ProcessedMessage) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AuditMixin{},
	}
}
