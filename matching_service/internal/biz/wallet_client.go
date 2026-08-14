package biz

import (
	"context"
	"errors"
	"log"

	"github.com/google/uuid"
)

type WalletClient interface {
	CheckBalance(ctx context.Context, userID uuid.UUID) (float64, error)
	HoldDeposit(ctx context.Context, userID uuid.UUID, amount float64) error
}

// 2. Tạo Mock (Hàng giả) để test trong Phase 1
type MockWalletClient struct{}

func NewMockWalletClient() WalletClient {
	return &MockWalletClient{}
}

// CheckBalance: Giả vờ như ông Tài xế/Chủ hàng nào cũng đang có 10 củ trong ví
func (m *MockWalletClient) CheckBalance(ctx context.Context, userID uuid.UUID) (float64, error) {
	log.Printf("[MockWallet] Check balance for user %s", userID)
	return 10000000.0, nil // 10 triệu VNĐ
}

// HoldDeposit: Hàm đóng băng tiền cọc
func (m *MockWalletClient) HoldDeposit(ctx context.Context, userID uuid.UUID, amount float64) error {
	log.Printf("[MockWallet] Freezing %f VND for user %s", amount, userID)

	// Giả lập logic bắt lỗi: Nếu đòi cọc lớn hơn 10 củ thì báo ví không đủ tiền
	if amount > 10000000.0 {
		return errors.New("insufficient balance in wallet to freeze escrow")
	}

	log.Println("[MockWallet] Freeze Escrow SUCCESS")
	return nil
}
