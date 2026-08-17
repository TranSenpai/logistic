package biz

import "context"

type UnitOfWorkRepository interface {
	Do(ctx context.Context, fn func(ctxTx context.Context) error) error
}
