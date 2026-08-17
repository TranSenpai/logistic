package repository

import (
	"context"
	"errors"
	"fmt"

	"wallet_service/ent"
)

type keyTx struct{}

type UnitOfWorkRepository interface {
	Do(ctx context.Context, fn func(ctxTx context.Context) error) error
}

type unitOfWorkRepository struct {
	entClient *ent.Client
}

func NewUnitOfWorkRepository(entClient *ent.Client) UnitOfWorkRepository {
	return &unitOfWorkRepository{entClient: entClient}
}

func (u *unitOfWorkRepository) Do(ctx context.Context, fn func(ctxTx context.Context) error) (err error) {
	tx, err := u.entClient.Tx(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			if errorRollBack := tx.Rollback(); errorRollBack != nil {
				err = errors.Join(err, fmt.Errorf("rollback err: %w", errorRollBack))
			}
		} else {
			if errorCommit := tx.Commit(); errorCommit != nil {
				err = fmt.Errorf("commit err: %w", errorCommit)
			}
		}
	}()

	ctxTx := context.WithValue(ctx, keyTx{}, tx.Client())
	err = fn(ctxTx)
	return
}

func GetClientTx(ctx context.Context, client *ent.Client) *ent.Client {
	clientAny := ctx.Value(keyTx{})
	if clientTx, ok := clientAny.(*ent.Client); ok {
		return clientTx
	}
	return client
}
