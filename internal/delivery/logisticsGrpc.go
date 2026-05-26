package delivery

import (
	"context"
	"fmt"
	"goBackend/internal/biz"
	logisticsv1 "goBackend/internal/dto/gen/go/logistics/v1"
	"goBackend/internal/entity"
	"time"
)

// LogisticsDelivery implements the gRPC interface defined in logistics.proto
// It acts as the boundary layer protecting the Biz logic.
type LogisticsDelivery struct {
	logisticsv1.UnimplementedMatchingEngineServiceServer
	matchingEngine *biz.MatchingEngine
}

func NewLogisticsDelivery(engine *biz.MatchingEngine) *LogisticsDelivery {
	return &LogisticsDelivery{
		matchingEngine: engine,
	}
}

// SubmitBid receives a cargo request from a shipper
func (d *LogisticsDelivery) SubmitBid(ctx context.Context, req *logisticsv1.SubmitBidRequest) (*logisticsv1.SubmitBidResponse, error) {
	// 1. Data Validation (Applying "100 Go Mistakes": Validate early and explicitly)
	if req.Origin == nil || req.Destination == nil {
		return nil, fmt.Errorf("origin and destination cannot be nil")
	}

	// 2. Mapping Protobuf to Domain Entity (Hexagonal Architecture principle)
	// We do not leak Protobuf structures into the Biz layer.
	bid := &entity.Bid{
		ID:       fmt.Sprintf("bid-%s-%d", req.UserId, time.Now().UnixNano()),
		UserID:   req.UserId,
		VolumeM3: req.VolumeM3,
		WeightKg: req.WeightKg,
		MaxPrice: req.MaxPrice,
		Origin: entity.Location{
			Latitude:  req.Origin.Latitude,
			Longitude: req.Origin.Longitude,
			ZoneID:    req.Origin.ZoneId,
		},
		Destination: entity.Location{
			Latitude:  req.Destination.Latitude,
			Longitude: req.Destination.Longitude,
			ZoneID:    req.Destination.ZoneId,
		},
		CreatedAt: time.Now(),
	}

	if req.ExpiresAt != nil && req.ExpiresAt.IsValid() {
		bid.ExpiresAt = req.ExpiresAt.AsTime()
	}

	// 3. Call Biz layer to execute core logic
	err := d.matchingEngine.SubmitBid(ctx, bid)
	if err != nil {
		return nil, err
	}

	return &logisticsv1.SubmitBidResponse{
		BidId:  bid.ID,
		Status: bid.Status,
	}, nil
}

// SubmitAsk receives an empty space report from a vehicle driver
func (d *LogisticsDelivery) SubmitAsk(ctx context.Context, req *logisticsv1.SubmitAskRequest) (*logisticsv1.SubmitAskResponse, error) {
	if req.CurrentLocation == nil || req.Destination == nil {
		return nil, fmt.Errorf("current_location and destination cannot be nil")
	}

	ask := &entity.Ask{
		ID:                fmt.Sprintf("ask-%s-%d", req.VehicleId, time.Now().UnixNano()),
		VehicleID:         req.VehicleId,
		DriverID:          req.DriverId,
		AvailableVolumeM3: req.AvailableVolumeM3,
		AvailableWeightKg: req.AvailableWeightKg,
		MinPrice:          req.MinPrice,
		CurrentLocation: entity.Location{
			Latitude:  req.CurrentLocation.Latitude,
			Longitude: req.CurrentLocation.Longitude,
			ZoneID:    req.CurrentLocation.ZoneId,
		},
		Destination: entity.Location{
			Latitude:  req.Destination.Latitude,
			Longitude: req.Destination.Longitude,
			ZoneID:    req.Destination.ZoneId,
		},
		CreatedAt: time.Now(),
	}

	if req.ExpiresAt != nil && req.ExpiresAt.IsValid() {
		ask.ExpiresAt = req.ExpiresAt.AsTime()
	}

	err := d.matchingEngine.SubmitAsk(ctx, ask)
	if err != nil {
		return nil, err
	}

	return &logisticsv1.SubmitAskResponse{
		AskId:  ask.ID,
		Status: ask.Status,
	}, nil
}
