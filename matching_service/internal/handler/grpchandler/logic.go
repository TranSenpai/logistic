package grpchandler

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	logisticsv1 "goBackend/api/logistic/matching/v1"
	"matching_service/internal/entity"
)

// SubmitBid receives a cargo request from a shipper
func (d *MatchingHandler) SubmitBid(ctx context.Context, req *logisticsv1.SubmitBidRequest) (*logisticsv1.SubmitBidResponse, error) {
	// 1. Data Validation (Applying "100 Go Mistakes": Validate early and explicitly)
	if req.Origin == nil || req.Destination == nil {
		return nil, fmt.Errorf("origin and destination cannot be nil")
	}

	userId, err := uuid.FromBytes(req.UserId)
	if err != nil {
		return nil, fmt.Errorf("invalid user_id: %w", err)
	}

	// 2. Mapping Protobuf to Domain Entity (Hexagonal Architecture principle)
	// We do not leak Protobuf structures into the Biz layer.
	bid := &entity.Bid{
		ID:       uuid.Must(uuid.NewV7()),
		UserID:   userId,
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
	matchedBid, err := d.matchingEngine.SubmitBid(ctx, bid)
	if err != nil {
		return nil, err
	}

	statusStr := "PENDING"
	if matchedBid.Status == entity.BidStatusMatched {
		statusStr = "MATCHED"
	}

	return &logisticsv1.SubmitBidResponse{
		BidId:  matchedBid.ID[:],
		Status: statusStr,
	}, nil
}

// SubmitAsk receives an empty space report from a vehicle driver
func (d *MatchingHandler) SubmitAsk(ctx context.Context, req *logisticsv1.SubmitAskRequest) (*logisticsv1.SubmitAskResponse, error) {
	if req.CurrentLocation == nil || req.Destination == nil {
		return nil, fmt.Errorf("current_location and destination cannot be nil")
	}

	vehicleId, err := uuid.FromBytes(req.VehicleId)
	if err != nil {
		return nil, fmt.Errorf("invalid vehicle_id: %w", err)
	}
	driverId, err := uuid.FromBytes(req.DriverId)
	if err != nil {
		return nil, fmt.Errorf("invalid driver_id: %w", err)
	}

	ask := &entity.Ask{
		ID:                uuid.Must(uuid.NewV7()),
		VehicleID:         vehicleId,
		DriverID:          driverId,
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

	matchedAsk, err := d.matchingEngine.SubmitAsk(ctx, ask)
	if err != nil {
		return nil, err
	}
	
	statusStr := "PENDING"
	if matchedAsk.Status == entity.BidStatusMatched { // assuming AskStatusMatched has same int mapping
		statusStr = "MATCHED"
	}

	return &logisticsv1.SubmitAskResponse{
		AskId:  matchedAsk.ID[:],
		Status: statusStr,
	}, nil
}
