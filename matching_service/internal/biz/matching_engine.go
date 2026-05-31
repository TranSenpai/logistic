package biz

import (
	"context"
	"log"
	"matching_service/internal/entity"
	"sync"
)

// MatchingEngine chịu trách nhiệm nhận Bid (Yêu cầu chở hàng) và Ask (Xe rỗng)
// Sau đó chạy thuật toán ghép nối chúng lại với nhau.
type MatchingEngine struct {
	mu      sync.RWMutex
	repo    IMatchingRepo
	spatial SpatialEngine
	// matchChan dùng để bắn kết quả ra ngoài sau khi ghép thành công
	matchChan chan *entity.MatchResult
}

func NewMatchingEngine(repo IMatchingRepo, spatial SpatialEngine) *MatchingEngine {
	return &MatchingEngine{
		repo:      repo,
		spatial:   spatial,
		matchChan: make(chan *entity.MatchResult, 1000),
	}
}

// SubmitBid nhận một yêu cầu chở hàng từ Shipper.
func (e *MatchingEngine) SubmitBid(ctx context.Context, bid *entity.Bid) error {
	if bid == nil {
		return entity.ErrNilBid
	}

	zoneID, err := e.spatial.GetZoneId(ctx, bid.Origin.Latitude, bid.Origin.Longitude)
	if err != nil {
		return err
	}
	bid.Origin.ZoneID = zoneID
	bid.Status = entity.BidStatusPending

	err = e.repo.CreateBid(ctx, bid)
	if err != nil {
		return err
	}

	e.matchForBid(ctx, bid)

	return nil
}

// SubmitAsk nhận thông tin xe rỗng từ Driver.
func (e *MatchingEngine) SubmitAsk(ctx context.Context, ask *entity.Ask) error {
	if ask == nil {
		return entity.ErrNilAsk
	}

	zoneID, err := e.spatial.GetZoneId(ctx, ask.CurrentLocation.Latitude, ask.CurrentLocation.Longitude)
	if err != nil {
		return err
	}
	ask.CurrentLocation.ZoneID = zoneID

	ask.Status = entity.AskStatusPending
	err = e.repo.CreateAsk(ctx, ask)
	if err != nil {
		return err
	}

	e.matchForAsk(ctx, ask)

	return nil
}

// matchForBid tìm kiếm các xe (Ask) phù hợp cho một đơn hàng (Bid) vừa được tạo.
func (e *MatchingEngine) matchForBid(ctx context.Context, bid *entity.Bid) {
	if bid == nil {
		log.Printf("Failed to find asks for bid: %v", entity.ErrNilBid)
		return
	}

	asks, err := e.repo.FindAskForBid(ctx, bid)
	if err != nil {
		log.Printf("Failed to find asks for bid %s: %v", bid.ID, err)
		return
	}

	if len(asks) > 0 {
		// TODO: Broadcast danh sách asks này cho các tài xế qua WebSocket/Push Notification
	}
}

// matchForAsk tìm kiếm các đơn hàng (Bid) phù hợp cho một xe (Ask) vừa được tạo.
func (e *MatchingEngine) matchForAsk(ctx context.Context, ask *entity.Ask) {
	if ask == nil {
		log.Printf("Failed to find bids for ask: %v", entity.ErrNilAsk)
		return
	}

	bids, err := e.repo.FindBidForAsk(ctx, ask)
	if err != nil {
		log.Printf("Failed to find bids for ask %s: %v", ask.ID, err)
		return
	}

	if len(bids) > 0 {
		// TODO: Broadcast danh sách bids này cho tài xế vừa tạo Ask
	}
}

// MatchStream trả về channel chứa kết quả ghép đơn thành công để Delivery/Worker hứng.
func (e *MatchingEngine) MatchStream() <-chan *entity.MatchResult {
	return e.matchChan
}
