package repository

import (
	"context"

	"github.com/google/uuid"
	"wallet_service/ent"
	"wallet_service/internal/entity"
	"wallet_service/internal/mapper"
)

type TransactionRepository interface {
	CreateTransaction(ctx context.Context, param *CreateTransactionParam) (*entity.Transaction, error)
}

type CreateTransactionParam struct {
	WalletID        uuid.UUID
	Amount          int64
	TransactionType uint8
	ReferenceID     string
	Description     string
	Status          uint8
}

type transactionRepository struct {
	client *ent.Client
	mapper mapper.WalletMapper
}

func NewTransactionRepository(client *ent.Client, m mapper.WalletMapper) TransactionRepository {
	return &transactionRepository{client: client, mapper: m}
}

func (r *transactionRepository) CreateTransaction(ctx context.Context, param *CreateTransactionParam) (*entity.Transaction, error) {
	client := GetClientTx(ctx, r.client)
	res, err := client.Transaction.Create().
		SetWalletID(param.WalletID).
		SetAmount(param.Amount).
		SetTransactionType(param.TransactionType).
		SetReferenceID(param.ReferenceID).
		SetDescription(param.Description).
		SetStatus(param.Status).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return r.mapper.EntToTransactionEntity(res), nil
}
