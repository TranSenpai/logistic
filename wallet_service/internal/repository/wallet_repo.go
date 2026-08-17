package repository

import (
	"context"

	"wallet_service/ent"
	"wallet_service/ent/wallet"
	"wallet_service/internal/entity"
	"wallet_service/internal/mapper"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
)

type WalletRepository interface {
	GetWalletForUpdate(ctx context.Context, userID uuid.UUID) (*entity.Wallet, error)
	GetWallet(ctx context.Context, userID uuid.UUID) (*entity.Wallet, error)
	UpdateBalance(ctx context.Context, walletID uuid.UUID, amount int64) error
	MarkMessageProcessed(ctx context.Context, messageID string) error
	CreateWallet(ctx context.Context, userID uuid.UUID, userType uint8) (*entity.Wallet, error)
}

type walletRepository struct {
	client *ent.Client
	mapper mapper.WalletMapper
}

func NewWalletRepository(client *ent.Client, m mapper.WalletMapper) WalletRepository {
	return &walletRepository{client: client, mapper: m}
}

func (r *walletRepository) MarkMessageProcessed(ctx context.Context, messageID string) error {
	client := GetClientTx(ctx, r.client)
	return client.ProcessedMessage.Create().SetID(messageID).Exec(ctx)
}

func (r *walletRepository) GetWalletForUpdate(ctx context.Context, userID uuid.UUID) (*entity.Wallet, error) {
	client := GetClientTx(ctx, r.client)
	res, err := client.Wallet.Query().
		Where(wallet.UserID(userID)).
		Modify(func(s *sql.Selector) {
			s.ForUpdate()
		}).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return r.mapper.EntToWalletEntity(res), nil
}

func (r *walletRepository) GetWallet(ctx context.Context, userID uuid.UUID) (*entity.Wallet, error) {
	client := GetClientTx(ctx, r.client)
	res, err := client.Wallet.Query().
		Where(wallet.UserID(userID)).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return r.mapper.EntToWalletEntity(res), nil
}

func (r *walletRepository) UpdateBalance(ctx context.Context, walletID uuid.UUID, amount int64) error {
	client := GetClientTx(ctx, r.client)
	return client.Wallet.UpdateOneID(walletID).
		AddBalance(amount).
		Exec(ctx)
}

func (r *walletRepository) CreateWallet(ctx context.Context, userID uuid.UUID, userType uint8) (*entity.Wallet, error) {
	client := GetClientTx(ctx, r.client)
	res, err := client.Wallet.Create().
		SetUserID(userID).
		SetUserType(userType).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return r.mapper.EntToWalletEntity(res), nil
}
