package softdelete

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// SoftDeleteMixin implements the soft delete pattern for schemas.
type SoftDeleteMixin struct {
	mixin.Schema
}

// Fields of the SoftDeleteMixin.
func (SoftDeleteMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("deleted_at").
			Optional(),
	}
}

type softDeleteKey struct{}

// SkipSoftDelete returns a new context that skips the soft-delete interceptor/mutators.
func SkipSoftDelete(parent context.Context) context.Context {
	return context.WithValue(parent, softDeleteKey{}, true)
}

// Interceptors of the SoftDeleteMixin.
func (d SoftDeleteMixin) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		// DÙNG Generic InterceptFunc thay vì intercept.TraverseFunc
		ent.InterceptFunc(func(next ent.Querier) ent.Querier {
			return ent.QuerierFunc(func(ctx context.Context, q ent.Query) (ent.Value, error) {
				// Skip soft-delete, means include soft-deleted entities.
				if skip, _ := ctx.Value(softDeleteKey{}).(bool); skip {
					return next.Query(ctx, q)
				}

				// Dùng Duck Typing để kiểm tra xem Query có hỗ trợ WhereP không
				if f, ok := q.(interface{ WhereP(...func(*sql.Selector)) }); ok {
					d.P(f)
				}
				return next.Query(ctx, q)
			})
		}),
	}
}

// Hooks of the SoftDeleteMixin.
func (d SoftDeleteMixin) Hooks() []ent.Hook {
	return []ent.Hook{
		func(next ent.Mutator) ent.Mutator {
			return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
				// Skip soft-delete, means delete the entity permanently.
				if skip, _ := ctx.Value(softDeleteKey{}).(bool); skip {
					return next.Mutate(ctx, m)
				}

				// 1. Xử lý cho luồng UPDATE: Chỉ chèn thêm điều kiện WHERE deleted_at IS NULL
				if m.Op().Is(ent.OpUpdateOne | ent.OpUpdate) {
					if f, ok := m.(interface{ WhereP(...func(*sql.Selector)) }); ok {
						d.P(f)
					}
					return next.Mutate(ctx, m)
				}

				// 2. Xử lý cho luồng DELETE: Đổi thành Update và gán deleted_at
				if m.Op().Is(ent.OpDeleteOne | ent.OpDelete) {
					if f, ok := m.(interface{ WhereP(...func(*sql.Selector)) }); ok {
						d.P(f)
					}

					// Dùng Duck Typing để lấy hàm SetOp mà không phụ thuộc vào generated code
					mx, ok := m.(interface{ SetOp(ent.Op) })
					if !ok {
						return nil, fmt.Errorf("unexpected mutation type %T", m)
					}

					mx.SetOp(ent.OpUpdate)
					err := m.SetField("deleted_at", time.Now())
					if err != nil {
						return nil, err
					}

					// KHÔNG GỌI mx.Client().Mutate() nữa.
					// Cứ đẩy tiếp cho chuỗi Mutator, nó sẽ tự động nhận diện đây là lệnh Update.
					return next.Mutate(ctx, m)
				}

				// Các thao tác khác (Create...) cứ đi qua bình thường
				return next.Mutate(ctx, m)
			})
		},
	}
}

// P adds a storage-level predicate to the queries and mutations.
func (d SoftDeleteMixin) P(w interface{ WhereP(...func(*sql.Selector)) }) {
	w.WhereP(
		sql.FieldIsNull(d.Fields()[0].Descriptor().Name),
	)
}
