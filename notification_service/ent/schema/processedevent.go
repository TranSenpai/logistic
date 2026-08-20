package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// ProcessedEvent là sổ chống xử lý trùng cho consumer RabbitMQ.
//
// RabbitMQ bảo đảm "ít nhất một lần", không phải "đúng một lần": chỉ cần service
// chết sau khi ghi DB mà chưa kịp ACK là message sẽ được giao lại. Không có bảng
// này thì tài xế nhận hai lần cùng một thông báo cho cùng một đơn.
//
// Cách dùng: INSERT event_id trong CÙNG transaction với việc tạo notification.
// Trùng khoá là dấu hiệu "đã xử lý rồi" -> ACK và bỏ qua.
type ProcessedEvent struct {
	ent.Schema
}

func (ProcessedEvent) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Unique(),
		field.String("event_id").Unique(),
		field.String("routing_key"),
		field.String("source").Optional(),
		field.Time("processed_at").Default(time.Now).Immutable(),
	}
}

func (ProcessedEvent) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("event_id").Unique(),
		index.Fields("processed_at"),
	}
}
