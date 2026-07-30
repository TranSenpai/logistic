package controller

import (
	"context"
	"fmt"
	"time"

	pb "github.com/logistic/api/logistic/matching_service/v1"

	"matching_service/internal/biz"
	"matching_service/internal/entity"
	"matching_service/internal/mapper"
	"matching_service/internal/mapper/generated"

	"github.com/google/uuid"
)

type matchingController struct {
	pb.UnimplementedMatchingEngineServiceServer
	matchingEngine biz.MatchingEngine
	mapper         mapper.AppMapper
}

func NewMatchingController(engine biz.MatchingEngine) *matchingController {
	return &matchingController{
		matchingEngine: engine,
		mapper:         &generated.AppMapperImpl{},
	}
}

// SubmitBid receives a cargo request from a shipper
func (c *matchingController) SubmitBid(ctx context.Context, req *pb.SubmitBidRequest) (*pb.SubmitBidResponse, error) {
	// 1. Data Validation (Applying "100 Go Mistakes": Validate early and explicitly)
	if req.Origin == nil || req.Destination == nil {
		return nil, fmt.Errorf("origin and destination cannot be nil")
	}

	// 2. Mapping Protobuf to Domain Entity (Hexagonal Architecture principle)
	bid, err := c.mapper.SubmitBidReqToEntity(req)
	if err != nil {
		return nil, fmt.Errorf("invalid request data: %w", err)
	}

	// Add generated fields that aren't mapped from request
	bid.ID = uuid.Must(uuid.NewV7())
	bid.CreatedAt = time.Now()

	// 3. Call Biz layer to execute core logic
	matchedBid, err := c.matchingEngine.SubmitBid(ctx, &bid)
	if err != nil {
		return nil, err
	}

	statusStr := "PENDING"
	if matchedBid.Status == entity.BidStatusMatched {
		statusStr = "MATCHED"
	}

	return &pb.SubmitBidResponse{
		BidId:  matchedBid.ID[:],
		Status: statusStr,
	}, nil
}

// SubmitAsk receives an empty space report from a vehicle driver
func (c *matchingController) SubmitAsk(ctx context.Context, req *pb.SubmitAskRequest) (*pb.SubmitAskResponse, error) {
	if req.CurrentLocation == nil || req.Destination == nil {
		return nil, fmt.Errorf("current_location and destination cannot be nil")
	}

	ask, err := c.mapper.SubmitAskReqToEntity(req)
	if err != nil {
		return nil, fmt.Errorf("invalid request data: %w", err)
	}

	// Add generated fields that aren't mapped from request
	ask.ID = uuid.Must(uuid.NewV7())
	ask.CreatedAt = time.Now()

	matchedAsk, err := c.matchingEngine.SubmitAsk(ctx, &ask)
	if err != nil {
		return nil, err
	}

	statusStr := "PENDING"
	if matchedAsk.Status == entity.BidStatusMatched { // assuming AskStatusMatched has same int mapping
		statusStr = "MATCHED"
	}

	return &pb.SubmitAskResponse{
		AskId:  matchedAsk.ID[:],
		Status: statusStr,
	}, nil
}
