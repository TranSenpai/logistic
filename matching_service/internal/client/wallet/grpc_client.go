package wallet

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	pb "github.com/logistic/api/logistic/wallet_service/v1"

	"matching_service/internal/biz"
)

// grpcClient là implementation hạ tầng của biz.WalletClient, gọi thật sang
// wallet_service qua gRPC. Nằm ngoài package biz để biz không phụ thuộc vào
// proto/gRPC — biz chỉ biết tới interface WalletClient.
type grpcClient struct {
	client pb.WalletServiceClient
}

func NewGrpcClient(client pb.WalletServiceClient) biz.WalletClient {
	return &grpcClient{client: client}
}

// CheckBalance trả về số dư theo đơn vị VND (float64) để so sánh trực tiếp
// với ConsensusDeposit. wallet_service lưu Balance ở đơn vị nhỏ nhất x100.
func (w *grpcClient) CheckBalance(ctx context.Context, userID uuid.UUID) (float64, error) {
	res, err := w.client.GetBalance(ctx, &pb.GetBalanceReq{UserId: userID[:]})
	if err != nil {
		return 0, fmt.Errorf("failed to get balance from wallet_service: %w", err)
	}
	return float64(res.Balance) / 100, nil
}
