package biz

import (
	"context"
	"matching_service/internal/entity"

	"github.com/google/uuid"
)

type MatchingRepo interface {
	CreateBid(ctx context.Context, bid *entity.Bid) error
	CreateAsk(ctx context.Context, ask *entity.Ask) error
	GetPendingBids(ctx context.Context, zone string) ([]entity.Bid, error)
	GetPendingAsks(ctx context.Context, zone string) ([]entity.Ask, error)
	FindAskForBid(ctx context.Context, bid *entity.Bid) ([]entity.Ask, error)
	FindBidForAsk(ctx context.Context, ask *entity.Ask) ([]entity.Bid, error)
	UpdateBid(ctx context.Context, bid *entity.Bid) error
	UpdateAsk(ctx context.Context, ask *entity.Ask) error
	DeleteAsk(ctx context.Context, id uuid.UUID) error
	DeleteBid(ctx context.Context, id uuid.UUID) error
}
