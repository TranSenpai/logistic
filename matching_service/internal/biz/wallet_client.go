package biz

import (
	"context"
	"log"

	"github.com/google/uuid"
)

// WalletClient kiểm tra số dư ví trước khi chốt match. Việc đóng băng tiền cọc
// (HoldDeposit) thực tế diễn ra bất đồng bộ qua Kafka topic "wallet.hold_deposit"
// (xem matchingEngineImpl.AcceptOffer), không qua interface này.
type WalletClient interface {
	CheckBalance(ctx context.Context, userID uuid.UUID) (float64, error)
}

// MockWalletClient dùng khi WALLET_SERVICE không cấu hình (vd. chạy test/local).
type MockWalletClient struct{}

func NewMockWalletClient() WalletClient {
	return &MockWalletClient{}
}

// CheckBalance: Giả vờ như ông Tài xế/Chủ hàng nào cũng đang có 10 củ trong ví
func (m *MockWalletClient) CheckBalance(ctx context.Context, userID uuid.UUID) (float64, error) {
	log.Printf("[MockWallet] Check balance for user %s", userID)
	return 10000000.0, nil // 10 triệu VNĐ
}
