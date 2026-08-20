package biz

import (
	"context"
	"fmt"

	"wallet_service/ent"
	"wallet_service/internal/entity"
	"wallet_service/internal/mapper"
	"wallet_service/internal/repository"
	"wallet_service/internal/search"

	"github.com/google/uuid"
)

var (
	SystemEscrowWalletID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
)

type WalletUseCase interface {
	GetBalance(ctx context.Context, userID uuid.UUID) (*entity.Wallet, error)
	Deposit(ctx context.Context, userID uuid.UUID, amount int64, description string) (*entity.Transaction, error)
	HoldDeposit(ctx context.Context, driverID uuid.UUID, amount int64, refID string) error
	ReleaseAndPay(ctx context.Context, driverID, shipperID uuid.UUID, escrowAmount, tripFare int64, refID string) error
	TransferMoney(ctx context.Context, fromUser, toUser uuid.UUID, amount int64, kafkaMsgID string) error
}

type wallet struct {
	uow        repository.UnitOfWorkRepository
	walletRepo repository.WalletRepository
	txRepo     repository.TransactionRepository
	esEngine   search.WalletSearchEngine
	mapper     mapper.WalletMapper
}

func NewWalletUseCase(
	uow repository.UnitOfWorkRepository,
	walletRepo repository.WalletRepository,
	txRepo repository.TransactionRepository,
	esEngine search.WalletSearchEngine,
	m mapper.WalletMapper,
) WalletUseCase {
	return &wallet{uow, walletRepo, txRepo, esEngine, m}
}

func (uc *wallet) GetBalance(ctx context.Context, userID uuid.UUID) (*entity.Wallet, error) {
	w, err := uc.walletRepo.GetWallet(ctx, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("%w: user %s", ErrWalletNotFound, userID)
		}
		return nil, fmt.Errorf("%w: %v", ErrDatabaseTxFailed, err)
	}
	return w, nil
}

func (uc *wallet) Deposit(ctx context.Context, userID uuid.UUID, amount int64, description string) (*entity.Transaction, error) {
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}

	var txRes *entity.Transaction
	var walletRes *entity.Wallet

	err := uc.uow.Do(ctx, func(ctxTx context.Context) error {
		w, err := uc.walletRepo.GetWalletForUpdate(ctxTx, userID)
		if err != nil {
			if ent.IsNotFound(err) {
				return fmt.Errorf("%w: user %s", ErrWalletNotFound, userID)
			}
			return fmt.Errorf("%w: %v", ErrDatabaseTxFailed, err)
		}

		if err := uc.walletRepo.UpdateBalance(ctxTx, w.ID, amount); err != nil {
			return err
		}

		tx, err := uc.txRepo.CreateTransaction(ctxTx, &repository.CreateTransactionParam{
			WalletID:        w.ID,
			Amount:          amount,
			TransactionType: 1,
			ReferenceID:     uuid.New().String(),
			Description:     description,
			Status:          1,
		})
		if err != nil {
			return err
		}

		walletRes, err = uc.walletRepo.GetWallet(ctxTx, userID)
		if err != nil {
			return err
		}
		txRes = tx
		return nil
	})

	if err == nil && uc.esEngine != nil {
		_ = uc.esEngine.IndexWallet(ctx, uc.mapper.EntityToESWallet(walletRes))
		_ = uc.esEngine.IndexTransaction(ctx, uc.mapper.EntityToESTransaction(txRes))
	}

	return txRes, err
}

func (uc *wallet) HoldDeposit(ctx context.Context, driverID uuid.UUID, amount int64, refID string) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}

	return uc.uow.Do(ctx, func(ctxTx context.Context) error {
		err := uc.walletRepo.MarkMessageProcessed(ctxTx, refID)
		if err != nil {
			if ent.IsConstraintError(err) {
				return nil
			}
			return fmt.Errorf("%w: %v", ErrDatabaseTxFailed, err)
		}

		driverWallet, err := uc.walletRepo.GetWalletForUpdate(ctxTx, driverID)
		if err != nil {
			if ent.IsNotFound(err) {
				return fmt.Errorf("%w: driver %s", ErrWalletNotFound, driverID)
			}
			return fmt.Errorf("%w: %v", ErrDatabaseTxFailed, err)
		}

		if driverWallet.Balance < amount {
			return fmt.Errorf("%w: required %d, has %d", ErrInsufficientBalance, amount, driverWallet.Balance)
		}

		escrowWallet, err := uc.walletRepo.GetWalletForUpdate(ctxTx, SystemEscrowWalletID)
		if err != nil {
			escrowWallet, err = uc.walletRepo.CreateWallet(ctxTx, SystemEscrowWalletID, 0)
			if err != nil {
				return err
			}
		}

		if err := uc.walletRepo.UpdateBalance(ctxTx, driverWallet.ID, -amount); err != nil {
			return err
		}
		if err := uc.walletRepo.UpdateBalance(ctxTx, escrowWallet.ID, amount); err != nil {
			return err
		}

		tx1, err := uc.txRepo.CreateTransaction(ctxTx, &repository.CreateTransactionParam{
			WalletID:        driverWallet.ID,
			Amount:          -amount,
			TransactionType: 5,
			ReferenceID:     refID,
			Description:     "Đóng băng tiền cọc nhận cuốc",
			Status:          1,
		})
		if err != nil {
			return err
		}

		tx2, err := uc.txRepo.CreateTransaction(ctxTx, &repository.CreateTransactionParam{
			WalletID:        escrowWallet.ID,
			Amount:          amount,
			TransactionType: 6,
			ReferenceID:     refID,
			Description:     "Hệ thống nhận tiền cọc của tài xế",
			Status:          1,
		})
		if err != nil {
			return err
		}

		if uc.esEngine != nil {
			w1, _ := uc.walletRepo.GetWallet(ctxTx, driverWallet.ID)
			w2, _ := uc.walletRepo.GetWallet(ctxTx, escrowWallet.ID)
			_ = uc.esEngine.IndexWallet(ctx, uc.mapper.EntityToESWallet(w1))
			_ = uc.esEngine.IndexWallet(ctx, uc.mapper.EntityToESWallet(w2))
			_ = uc.esEngine.IndexTransaction(ctx, uc.mapper.EntityToESTransaction(tx1))
			_ = uc.esEngine.IndexTransaction(ctx, uc.mapper.EntityToESTransaction(tx2))
		}
		return nil
	})
}

func (uc *wallet) ReleaseAndPay(ctx context.Context, driverID, shipperID uuid.UUID, escrowAmount, tripFare int64, refID string) error {
	return uc.uow.Do(ctx, func(ctxTx context.Context) error {
		err := uc.walletRepo.MarkMessageProcessed(ctxTx, refID)
		if err != nil {
			if ent.IsConstraintError(err) {
				return nil
			}
			return err
		}

		escrowWallet, err := uc.walletRepo.GetWalletForUpdate(ctxTx, SystemEscrowWalletID)
		if err != nil {
			if ent.IsNotFound(err) {
				return ErrSystemWalletNotFound
			}
			return fmt.Errorf("%w: %v", ErrDatabaseTxFailed, err)
		}
		if escrowWallet.Balance < escrowAmount {
			return fmt.Errorf("%w: system escrow required %d, has %d", ErrInsufficientBalance, escrowAmount, escrowWallet.Balance)
		}

		shipperWallet, err := uc.walletRepo.GetWalletForUpdate(ctxTx, shipperID)
		if err != nil {
			if ent.IsNotFound(err) {
				return fmt.Errorf("%w: shipper %s", ErrWalletNotFound, shipperID)
			}
			return fmt.Errorf("%w: %v", ErrDatabaseTxFailed, err)
		}
		if shipperWallet.Balance < tripFare {
			return fmt.Errorf("%w: shipper required %d, has %d", ErrInsufficientBalance, tripFare, shipperWallet.Balance)
		}

		driverWallet, err := uc.walletRepo.GetWalletForUpdate(ctxTx, driverID)
		if err != nil {
			if ent.IsNotFound(err) {
				return fmt.Errorf("%w: driver %s", ErrWalletNotFound, driverID)
			}
			return fmt.Errorf("%w: %v", ErrDatabaseTxFailed, err)
		}

		if err := uc.walletRepo.UpdateBalance(ctxTx, escrowWallet.ID, -escrowAmount); err != nil {
			return err
		}
		if err := uc.walletRepo.UpdateBalance(ctxTx, driverWallet.ID, escrowAmount); err != nil {
			return err
		}

		if err := uc.walletRepo.UpdateBalance(ctxTx, shipperWallet.ID, -tripFare); err != nil {
			return err
		}
		if err := uc.walletRepo.UpdateBalance(ctxTx, driverWallet.ID, tripFare); err != nil {
			return err
		}

		tx1, err := uc.txRepo.CreateTransaction(ctxTx, &repository.CreateTransactionParam{
			WalletID: escrowWallet.ID, Amount: -escrowAmount, TransactionType: 7, ReferenceID: refID, Description: "Hoàn cọc cho tài xế", Status: 1,
		})
		if err != nil {
			return err
		}

		tx2, err := uc.txRepo.CreateTransaction(ctxTx, &repository.CreateTransactionParam{
			WalletID: driverWallet.ID, Amount: escrowAmount, TransactionType: 8, ReferenceID: refID, Description: "Nhận lại tiền cọc", Status: 1,
		})
		if err != nil {
			return err
		}

		tx3, err := uc.txRepo.CreateTransaction(ctxTx, &repository.CreateTransactionParam{
			WalletID: shipperWallet.ID, Amount: -tripFare, TransactionType: 4, ReferenceID: refID, Description: "Thanh toán phí cuốc xe", Status: 1,
		})
		if err != nil {
			return err
		}

		tx4, err := uc.txRepo.CreateTransaction(ctxTx, &repository.CreateTransactionParam{
			WalletID: driverWallet.ID, Amount: tripFare, TransactionType: 3, ReferenceID: refID, Description: "Nhận cước phí cuốc xe", Status: 1,
		})
		if err != nil {
			return err
		}

		if uc.esEngine != nil {
			w1, _ := uc.walletRepo.GetWallet(ctxTx, escrowWallet.ID)
			w2, _ := uc.walletRepo.GetWallet(ctxTx, driverWallet.ID)
			w3, _ := uc.walletRepo.GetWallet(ctxTx, shipperWallet.ID)
			_ = uc.esEngine.IndexWallet(ctx, uc.mapper.EntityToESWallet(w1))
			_ = uc.esEngine.IndexWallet(ctx, uc.mapper.EntityToESWallet(w2))
			_ = uc.esEngine.IndexWallet(ctx, uc.mapper.EntityToESWallet(w3))
			_ = uc.esEngine.IndexTransaction(ctx, uc.mapper.EntityToESTransaction(tx1))
			_ = uc.esEngine.IndexTransaction(ctx, uc.mapper.EntityToESTransaction(tx2))
			_ = uc.esEngine.IndexTransaction(ctx, uc.mapper.EntityToESTransaction(tx3))
			_ = uc.esEngine.IndexTransaction(ctx, uc.mapper.EntityToESTransaction(tx4))
		}

		return nil
	})
}

func (w *wallet) TransferMoney(ctx context.Context, fromUser, toUser uuid.UUID, amount int64, kafkaMsgID string) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if fromUser == toUser {
		return ErrSelfTransfer
	}

	return w.uow.Do(ctx, func(ctxTx context.Context) error {
		err := w.walletRepo.MarkMessageProcessed(ctxTx, kafkaMsgID)
		if err != nil {
			if ent.IsConstraintError(err) {
				return nil
			}
			return fmt.Errorf("%w: %v", ErrDatabaseTxFailed, err)
		}

		senderWallet, err := w.walletRepo.GetWalletForUpdate(ctxTx, fromUser)
		if err != nil {
			if ent.IsNotFound(err) {
				return fmt.Errorf("%w: sender %s", ErrWalletNotFound, fromUser)
			}
			return fmt.Errorf("%w: %v", ErrDatabaseTxFailed, err)
		}
		if senderWallet.Balance < amount {
			return fmt.Errorf("%w: sender required %d, has %d", ErrInsufficientBalance, amount, senderWallet.Balance)
		}

		receiverWallet, err := w.walletRepo.GetWalletForUpdate(ctxTx, toUser)
		if err != nil {
			if ent.IsNotFound(err) {
				return fmt.Errorf("%w: receiver %s", ErrWalletNotFound, toUser)
			}
			return fmt.Errorf("%w: %v", ErrDatabaseTxFailed, err)
		}

		if err := w.walletRepo.UpdateBalance(ctxTx, senderWallet.ID, -amount); err != nil {
			return err
		}
		if err := w.walletRepo.UpdateBalance(ctxTx, receiverWallet.ID, amount); err != nil {
			return err
		}

		tx1, err := w.txRepo.CreateTransaction(ctxTx, &repository.CreateTransactionParam{
			WalletID:        senderWallet.ID,
			Amount:          -amount,
			TransactionType: 3,
			ReferenceID:     kafkaMsgID,
			Description:     "Chuyển tiền",
			Status:          1,
		})
		if err != nil {
			return err
		}

		tx2, err := w.txRepo.CreateTransaction(ctxTx, &repository.CreateTransactionParam{
			WalletID:        receiverWallet.ID,
			Amount:          amount,
			TransactionType: 4,
			ReferenceID:     kafkaMsgID,
			Description:     "Nhận tiền",
			Status:          1,
		})
		if err != nil {
			return err
		}

		if w.esEngine != nil {
			w1, _ := w.walletRepo.GetWallet(ctxTx, senderWallet.ID)
			w2, _ := w.walletRepo.GetWallet(ctxTx, receiverWallet.ID)
			_ = w.esEngine.IndexWallet(ctx, w.mapper.EntityToESWallet(w1))
			_ = w.esEngine.IndexWallet(ctx, w.mapper.EntityToESWallet(w2))
			_ = w.esEngine.IndexTransaction(ctx, w.mapper.EntityToESTransaction(tx1))
			_ = w.esEngine.IndexTransaction(ctx, w.mapper.EntityToESTransaction(tx2))
		}

		return nil
	})
}