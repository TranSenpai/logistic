package biz

import (
	"context"
	"fmt"
	"log"
	cerr "matching_service/internal/common/errors"
	"matching_service/internal/entity"
	"sync"

	"github.com/google/uuid"
)

// MatchingEngine chịu trách nhiệm nhận Bid (Yêu cầu chở hàng) và Ask (Xe rỗng)
// Sau đó chạy thuật toán ghép nối chúng lại với nhau.
type MatchingEngine interface {
	SubmitBid(ctx context.Context, bid *entity.Bid) (*entity.Bid, error)
	SubmitAsk(ctx context.Context, ask *entity.Ask) (*entity.Ask, error)
	MatchStream() <-chan *entity.MatchResult
}

type matchingEngineImpl struct {
	mu      sync.RWMutex
	repo    MatchingRepo
	spatial SpatialEngine
	// matchChan dùng để bắn kết quả ra ngoài sau khi ghép thành công
	matchChan chan *entity.MatchResult
	kafkaPub  EventPublisher
	natsPub   EventPublisher
}

func NewMatchingEngine(repo MatchingRepo, spatial SpatialEngine, kafkaPub EventPublisher, natsPub EventPublisher) MatchingEngine {
	return &matchingEngineImpl{
		repo:      repo,
		spatial:   spatial,
		kafkaPub:  kafkaPub,
		natsPub:   natsPub,
		matchChan: make(chan *entity.MatchResult, 1000),
	}
}

// SubmitBid nhận một yêu cầu chở hàng từ Shipper.
func (e *matchingEngineImpl) SubmitBid(ctx context.Context, bid *entity.Bid) (*entity.Bid, error) {
	if bid == nil {
		return nil, fmt.Errorf("%w: %v", cerr.ErrInvalidID, "bid is nil")
	}

	zoneID, err := e.spatial.GetZoneId(ctx, bid.Origin.Latitude, bid.Origin.Longitude)
	if err != nil {
		return nil, err
	}
	bid.Origin.ZoneID = zoneID
	bid.Status = entity.BidStatusPending

	if bid.ID == uuid.Nil {
		bid.ID = uuid.Must(uuid.NewV7())
	}

	err = e.repo.CreateBid(ctx, bid)
	if err != nil {
		return nil, err
	}

	e.matchForBid(ctx, bid)

	return bid, nil
}

// SubmitAsk nhận thông tin xe rỗng từ Driver.
func (e *matchingEngineImpl) SubmitAsk(ctx context.Context, ask *entity.Ask) (*entity.Ask, error) {
	if ask == nil {
		return nil, fmt.Errorf("%w: %v", cerr.ErrInvalidID, "ask is nil")
	}

	zoneID, err := e.spatial.GetZoneId(ctx, ask.CurrentLocation.Latitude, ask.CurrentLocation.Longitude)
	if err != nil {
		return nil, err
	}
	ask.CurrentLocation.ZoneID = zoneID
	ask.Status = entity.AskStatusPending

	if ask.ID == uuid.Nil {
		ask.ID = uuid.Must(uuid.NewV7())
	}

	err = e.repo.CreateAsk(ctx, ask)
	if err != nil {
		return nil, err
	}

	e.matchForAsk(ctx, ask)

	return ask, nil
}

// matchForBid tìm kiếm các xe (Ask) phù hợp cho một đơn hàng (Bid) vừa được tạo.
func (e *matchingEngineImpl) matchForBid(ctx context.Context, bid *entity.Bid) {
	if bid == nil {
		return
	}

	asks, err := e.repo.FindAskForBid(ctx, bid)
	if err != nil {
		log.Printf("Failed to find asks for bid %s: %v", bid.ID, err)
		return
	}

	if len(asks) == 0 {
		return
	}

	// Basic Rule-based Scoring Algorithm (MVP Stage 1)
	// Trọng số: Giá cả (50%), Lấp đầy thể tích (50%)
	type ScoredAsk struct {
		Ask   entity.Ask
		Score float64
	}

	var scoredAsks []ScoredAsk
	for _, ask := range asks {
		score := 0.0

		// 1. Tiêu chí giá cả (Max 50 điểm)
		// Nếu Ask MinPrice thấp hơn hoặc bằng Bid MaxPrice -> Có thể khớp
		priceDiff := bid.MaxPrice - ask.MinPrice
		if priceDiff >= 0 {
			// Càng chênh lệch nhiều (tài xế chịu giá rẻ hơn) -> Điểm càng cao
			score += 30.0 + (priceDiff/bid.MaxPrice)*20.0
		}

		// 2. Tiêu chí lấp đầy thể tích (Max 50 điểm)
		if ask.AvailableVolumeM3 > 0 {
			volumeRatio := bid.VolumeM3 / ask.AvailableVolumeM3
			score += volumeRatio * 50.0
		}

		scoredAsks = append(scoredAsks, ScoredAsk{Ask: ask, Score: score})
	}
	e.natsPub.Publish(ctx, &EventMessage{
		Topic:   "matching.drivers.notified",
		Key:     bid.ID.String(),
		Payload: scoredAsks,
	})

	log.Printf("Found %d potential drivers for Bid %s. Top score: %.2f", len(scoredAsks), bid.ID, scoredAsks[0].Score)
}

// matchForAsk tìm kiếm các đơn hàng (Bid) phù hợp cho một xe (Ask) vừa được tạo.
func (e *matchingEngineImpl) matchForAsk(ctx context.Context, ask *entity.Ask) {
	if ask == nil {
		return
	}

	bids, err := e.repo.FindBidForAsk(ctx, ask)
	if err != nil {
		log.Printf("Failed to find bids for ask %s: %v", ask.ID, err)
		return
	}

	if len(bids) > 0 {
		// TODO: Tương tự như trên, chấm điểm ngược lại cho mảng Bids
		log.Printf("Found %d potential bids for Ask %s", len(bids), ask.ID)
	}
}

// MatchStream trả về channel chứa kết quả ghép đơn thành công để Delivery/Worker hứng.
func (e *matchingEngineImpl) MatchStream() <-chan *entity.MatchResult {
	return e.matchChan
}
