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

type MatchingEngine interface {
	SubmitBid(ctx context.Context, bid *entity.Bid) (*entity.Bid, error)
	SubmitAsk(ctx context.Context, ask *entity.Ask) (*entity.Ask, error)
	SubmitOffer(ctx context.Context, bidID uuid.UUID, askID uuid.UUID, desiredPrice float64) error
	AcceptOffer(ctx context.Context, bidID uuid.UUID, askID uuid.UUID) (*entity.MatchContract, error)
	RejectOffer(ctx context.Context, bidID uuid.UUID, askID uuid.UUID) error
	MatchStream() <-chan *entity.MatchContract
}

type matchingEngineImpl struct {
	repo         MatchingRepo
	spatial      SpatialEngine
	walletClient WalletClient
	matchChan    chan *entity.MatchContract
	kafkaPub     EventPublisher
	natsPub      EventPublisher
	mu           sync.RWMutex
}

func NewMatchingEngine(repo MatchingRepo, spatial SpatialEngine, walletClient WalletClient, kafkaPub EventPublisher, natsPub EventPublisher) MatchingEngine {
	return &matchingEngineImpl{
		repo:         repo,
		spatial:      spatial,
		walletClient: walletClient,
		kafkaPub:     kafkaPub,
		natsPub:      natsPub,
		matchChan:    make(chan *entity.MatchContract, 1000),
	}
}

func (e *matchingEngineImpl) broadcastBidToDrivers(ctx context.Context, bid *entity.Bid) {
	if bid == nil {
		return
	}

	asks, err := e.repo.FindAskForBid(ctx, bid)
	if err != nil {
		log.Printf("Failed to find asks for bid %s: %v", bid.ID, err)
		return
	}

	if len(asks) == 0 {
		log.Printf("No available drivers found for Bid %s", bid.ID)
		return
	}

	e.natsPub.Publish(ctx, &EventMessage{
		Topic:   "matching.drivers.notified",
		Key:     bid.ID.String(),
		Payload: asks,
	})

	e.kafkaPub.Publish(ctx, &EventMessage{
		Topic:   "matching.bid_created",
		Key:     bid.ID.String(),
		Payload: *bid,
	})

	log.Printf("[BROADCAST] Found %d potential drivers for Bid %s", len(asks), bid.ID)
}

func (e *matchingEngineImpl) suggestBidsToDriver(ctx context.Context, ask *entity.Ask) {
	if ask == nil {
		return
	}

	bids, err := e.repo.FindBidForAsk(ctx, ask)
	if err != nil {
		log.Printf("Failed to find bids for ask %s: %v", ask.ID, err)
		return
	}

	if len(bids) > 0 {
		e.natsPub.Publish(ctx, &EventMessage{
			Topic:   fmt.Sprintf("matching.suggested_cargos.%s", ask.DriverID.String()),
			Key:     ask.ID.String(),
			Payload: bids,
		})

		e.kafkaPub.Publish(ctx, &EventMessage{
			Topic:   "matching.ask_created",
			Key:     ask.ID.String(),
			Payload: *ask,
		})

		log.Printf("[SUGGESTION] Found %d pending bids for Driver %s", len(bids), ask.DriverID)
	}
}

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

	e.broadcastBidToDrivers(ctx, bid)

	return bid, nil
}

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

	e.suggestBidsToDriver(ctx, ask)

	return ask, nil
}

func (e *matchingEngineImpl) MatchStream() <-chan *entity.MatchContract {
	return e.matchChan
}

func (e *matchingEngineImpl) SubmitOffer(ctx context.Context, bidID uuid.UUID, askID uuid.UUID, desiredPrice float64) error {
	ask, err := e.repo.GetAsk(ctx, askID)
	if err != nil {
		return fmt.Errorf("failed to retrieve ask: %w", err)
	}

	ask.MinPrice = desiredPrice
	topic := fmt.Sprintf("matching.offers.%s", bidID.String())

	err = e.natsPub.Publish(ctx, &EventMessage{
		Topic:   topic,
		Key:     ask.ID.String(),
		Payload: *ask,
	})
	if err != nil {
		return fmt.Errorf("failed to publish offer: %w", err)
	}

	log.Printf("[OFFER SUBMITTED] Driver %s -> Bid %s. Price: %.2f", ask.DriverID, bidID, desiredPrice)
	return nil
}

func (e *matchingEngineImpl) ProcessOfferQueue(ctx context.Context, bidID uuid.UUID, offerAsk *entity.Ask) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	bid, err := e.repo.GetBid(ctx, bidID)
	if err != nil {
		return err
	}

	if bid.Status != entity.BidStatusPending {
		e.natsPub.Publish(ctx, &EventMessage{
			Topic:   fmt.Sprintf("matching.drivers.rejected.%s", offerAsk.DriverID.String()),
			Key:     bidID.String(),
			Payload: []byte("Bid is already under negotiation."),
		})
		log.Printf("[INSTANT RELEASE] Driver %s rejected for Bid %s (Already taken)", offerAsk.DriverID, bidID)
		return nil
	}

	bid.Status = entity.BidStatusNegotiating
	if err := e.repo.UpdateBid(ctx, bid); err != nil {
		return err
	}

	e.natsPub.Publish(ctx, &EventMessage{
		Topic:   fmt.Sprintf("matching.shipper.offers_received.%s", bid.ShipperID.String()),
		Key:     offerAsk.ID.String(),
		Payload: *offerAsk,
	})

	log.Printf("[OFFER FORWARDED] Driver %s sent to Shipper %s for Bid %s", offerAsk.DriverID, bid.ShipperID, bidID)

	return nil // ACK message
}

func (e *matchingEngineImpl) RejectOffer(ctx context.Context, bidID uuid.UUID, askID uuid.UUID) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	bid, err := e.repo.GetBid(ctx, bidID)
	if err != nil {
		return err
	}

	if bid.Status != entity.BidStatusNegotiating {
		return fmt.Errorf("cannot reject: bid is not in negotiating status")
	}

	bid.Status = entity.BidStatusPending
	if err := e.repo.UpdateBid(ctx, bid); err != nil {
		return err
	}

	ask, err := e.repo.GetAsk(ctx, askID)
	if err == nil {
		e.natsPub.Publish(ctx, &EventMessage{
			Topic:   fmt.Sprintf("matching.drivers.rejected.%s", ask.DriverID.String()),
			Key:     bidID.String(),
			Payload: []byte("Shipper has rejected your offer."),
		})
	}

	log.Printf("[OFFER REJECTED] Shipper %s rejected Driver %s for Bid %s", bid.ShipperID, ask.DriverID, bidID)
	return nil
}

func (e *matchingEngineImpl) AcceptOffer(ctx context.Context, bidID uuid.UUID, askID uuid.UUID) (*entity.MatchContract, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	bid, err := e.repo.GetBid(ctx, bidID)
	if err != nil {
		return nil, err
	}

	if bid.Status != entity.BidStatusNegotiating {
		return nil, fmt.Errorf("cannot accept: bid is not in negotiating status")
	}

	ask, err := e.repo.GetAsk(ctx, askID)
	if err != nil {
		return nil, err
	}

	contract := &entity.MatchContract{
		ID:               uuid.Must(uuid.NewV7()),
		BidID:            bid.ID,
		AskID:            ask.ID,
		ConsensusPrice:   ask.MinPrice,
		ConsensusDeposit: ask.MinPrice * 0.1, // Escrow deposit: 10%
		Status:           entity.MatchStatusAccepted,
	}

	if e.walletClient != nil {
		if err := e.walletClient.HoldDeposit(ctx, bid.ShipperID, contract.ConsensusDeposit); err != nil {
			return nil, fmt.Errorf("failed to freeze shipper deposit: %w", err)
		}
	}

	bid.Status = entity.BidStatusMatched
	ask.Status = entity.AskStatusMatched

	if err := e.repo.UpdateBid(ctx, bid); err != nil {
		return nil, err
	}
	if err := e.repo.UpdateAsk(ctx, ask); err != nil {
		return nil, err
	}
	if err := e.repo.CreateMatchContract(ctx, contract); err != nil {
		return nil, err
	}

	select {
	case e.matchChan <- contract:
	default:
		log.Println("WARNING: matchChan is full!")
	}

	e.natsPub.Publish(ctx, &EventMessage{
		Topic:   fmt.Sprintf("matching.drivers.accepted.%s", ask.DriverID.String()),
		Key:     contract.ID.String(),
		Payload: *contract,
	})

	log.Printf("[DEAL CLOSED] MatchContract %s created! Shipper: %s, Driver: %s, Price: %.2f",
		contract.ID, bid.ShipperID, ask.DriverID, contract.ConsensusPrice)

	return contract, nil
}
