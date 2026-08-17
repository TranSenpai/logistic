package repo

import (
	"context"
	"errors"
	"fmt"
	"user_service/ent"
	"wallet_service/internal/biz"
)

type keyTx struct{}

type UnitOfWork struct {
	entClient *ent.Client
}

func NewUnitOfWorkRepository(entClient *ent.Client) biz.UnitOfWorkRepository {
	return &UnitOfWork{
		entClient: entClient,
	}
}

func (u *UnitOfWork) Do(ctx context.Context, fn func(ctxTx context.Context) error) (err error) {
	tx, err := u.entClient.Tx(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			// rollback
			errorRollback := tx.Rollback()
			if errorRollback != nil {
				fmt.Println("errorRollBack", errorRollback)
				err = errors.New("rolback err")
			}
		} else {
			// commit
			errorCommit := tx.Commit()
			if errorCommit != nil {
				fmt.Println("errorCommit", errorCommit)
				err = errors.New("commit err")
			}

		}
	}()

	ctxTx := context.WithValue(ctx, keyTx{}, tx.Client())
	err = fn(ctxTx)
	if err != nil {
		return err
	}

	return nil
}

func GetClientTx(ctx context.Context, client *ent.Client) *ent.Client {
	clientAny := ctx.Value(keyTx{})

	// Check xem client có đang trong transaction không
	clientTx, ok := clientAny.(*ent.Client)
	if !ok {
		return client
	}

	return clientTx
}
