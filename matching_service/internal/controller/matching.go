package controller

import (
	"context"
	"fmt"

	pb "github.com/logistic/api/logistic/matching_service/v1"

	"matching_service/internal/biz"
	"matching_service/internal/mapper"
	"matching_service/internal/mapper/generated"
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

func (c *matchingController) SubmitBid(ctx context.Context, req *pb.SubmitBidRequest) (*pb.SubmitBidResponse, error) {
	if req.Payload == nil || req.Payload.Origin == nil || req.Payload.Destination == nil {
		return nil, fmt.Errorf("payload, origin, and destination cannot be nil")
	}

	bid, err := c.mapper.PbBidToEntity(req.Payload)
	if err != nil {
		return nil, fmt.Errorf("invalid request data: %w", err)
	}

	matchedBid, err := c.matchingEngine.SubmitBid(ctx, &bid)
	if err != nil {
		return nil, err
	}

	return &pb.SubmitBidResponse{
		BidId:  matchedBid.ID[:],
		Status: matchedBid.StatusString(),
	}, nil
}

func (c *matchingController) SubmitAsk(ctx context.Context, req *pb.SubmitAskRequest) (*pb.SubmitAskResponse, error) {
	if req.Payload == nil || req.Payload.CurrentLocation == nil || req.Payload.Destination == nil {
		return nil, fmt.Errorf("payload, current_location, and destination cannot be nil")
	}

	ask, err := c.mapper.PbAskToEntity(req.Payload)
	if err != nil {
		return nil, fmt.Errorf("invalid request data: %w", err)
	}

	matchedAsk, err := c.matchingEngine.SubmitAsk(ctx, &ask)
	if err != nil {
		return nil, err
	}

	return &pb.SubmitAskResponse{
		AskId:  matchedAsk.ID[:],
		Status: matchedAsk.StatusString(),
	}, nil
}
