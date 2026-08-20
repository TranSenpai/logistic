package controller

import (
	"gateway_service/internal/middleware"
	"gateway_service/internal/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	pb "github.com/logistic/api/logistic/matching_service/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MatchingController struct {
	matchingClient pb.MatchingEngineServiceClient
}

func NewMatchingController(matchingClient pb.MatchingEngineServiceClient) *MatchingController {
	return &MatchingController{matchingClient: matchingClient}
}

func uuidBytes(raw string) ([]byte, bool) {
	if raw == "" {
		return nil, true
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil, false
	}
	return id[:], true
}

type LocationReq struct {
	Latitude  float64 `json:"latitude" binding:"required,latitude"`
	Longitude float64 `json:"longitude" binding:"required,longitude"`
	ZoneID    string  `json:"zone_id"`
}

func (l *LocationReq) toPb() *pb.Location {
	if l == nil {
		return nil
	}
	return &pb.Location{Latitude: l.Latitude, Longitude: l.Longitude, ZoneId: l.ZoneID}
}

type SubmitBidReq struct {
	ShipperID       string       `json:"shipper_id" binding:"required"`
	ShipperPhone    string       `json:"shipper_phone"`
	ShipperMail     string       `json:"shipper_mail"`
	ConsigneeID     string       `json:"consignee_id"`
	ConsigneePhone  string       `json:"consignee_phone"`
	ConsigneeMail   string       `json:"consignee_mail"`
	Origin          *LocationReq `json:"origin" binding:"required"`
	Destination     *LocationReq `json:"destination" binding:"required"`
	VolumeM3        float64      `json:"volume_m3" binding:"required,gt=0"`
	WeightKg        float64      `json:"weight_kg" binding:"required,gt=0"`
	MaxPrice        float64      `json:"max_price" binding:"required,gt=0"`
	CargoValue      float64      `json:"cargo_value"`
	RequiredDeposit float64      `json:"required_deposit"`
	DesiredDeposit  float64      `json:"desired_deposit"`
}

// SubmitBid godoc
// @Summary      Chủ hàng đăng đơn cần xe
// @Description  Sau khi lưu đơn, matching_service tìm tài xế phù hợp và phát sự kiện qua RabbitMQ để notification_service báo cho từng tài xế.
// @Tags         Matching
// @Accept       json
// @Produce      json
// @Success      200 {object} response.Envelope
// @Router       /api/v1/matching/bids [post]
func (c *MatchingController) SubmitBid(ctx *gin.Context) {
	var req SubmitBidReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "VALIDATION_FAILED", err.Error())
		return
	}

	shipperID, ok := uuidBytes(req.ShipperID)
	if !ok {
		response.BadRequest(ctx, "INVALID_SHIPPER_ID", "shipper_id phải là UUID hợp lệ")
		return
	}
	consigneeID, ok := uuidBytes(req.ConsigneeID)
	if !ok {
		response.BadRequest(ctx, "INVALID_CONSIGNEE_ID", "consignee_id phải là UUID hợp lệ")
		return
	}

	resp, err := c.matchingClient.SubmitBid(ctx.Request.Context(), &pb.SubmitBidRequest{
		RequestedAt: timestamppb.Now(),
		Payload: &pb.Bid{
			ShipperId:       shipperID,
			ShipperPhone:    req.ShipperPhone,
			ShipperMail:     req.ShipperMail,
			ConsigneeId:     consigneeID,
			ConsigneePhone:  req.ConsigneePhone,
			ConsigneeMail:   req.ConsigneeMail,
			Origin:          req.Origin.toPb(),
			Destination:     req.Destination.toPb(),
			VolumeM3:        req.VolumeM3,
			WeightKg:        req.WeightKg,
			MaxPrice:        req.MaxPrice,
			CargoValue:      req.CargoValue,
			RequiredDeposit: req.RequiredDeposit,
			DesiredDeposit:  req.DesiredDeposit,
		},
	})
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.OK(ctx, gin.H{
		"bid_id": bytesToUUIDString(resp.BidId),
		"status": resp.Status,
	})
}

type SubmitAskReq struct {
	DriverID          string       `json:"driver_id" binding:"required"`
	DriverPhone       string       `json:"driver_phone"`
	DriverMail        string       `json:"driver_mail"`
	VehicleID         string       `json:"vehicle_id" binding:"required"`
	CurrentLocation   *LocationReq `json:"current_location" binding:"required"`
	Destination       *LocationReq `json:"destination" binding:"required"`
	AvailableVolumeM3 float64      `json:"available_volume_m3" binding:"required,gt=0"`
	AvailableWeightKg float64      `json:"available_weight_kg" binding:"required,gt=0"`
	MinPrice          float64      `json:"min_price" binding:"required,gt=0"`
	DesiredDeposit    float64      `json:"desired_deposit"`
}

// SubmitAsk godoc
// @Summary      Tài xế đăng chuyến còn chỗ trống
// @Tags         Matching
// @Accept       json
// @Produce      json
// @Success      200 {object} response.Envelope
// @Router       /api/v1/matching/asks [post]
func (c *MatchingController) SubmitAsk(ctx *gin.Context) {
	var req SubmitAskReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "VALIDATION_FAILED", err.Error())
		return
	}

	driverID, ok := uuidBytes(req.DriverID)
	if !ok {
		response.BadRequest(ctx, "INVALID_DRIVER_ID", "driver_id phải là UUID hợp lệ")
		return
	}
	vehicleID, ok := uuidBytes(req.VehicleID)
	if !ok {
		response.BadRequest(ctx, "INVALID_VEHICLE_ID", "vehicle_id phải là UUID hợp lệ")
		return
	}

	resp, err := c.matchingClient.SubmitAsk(ctx.Request.Context(), &pb.SubmitAskRequest{
		RequestedAt: timestamppb.Now(),
		Payload: &pb.Ask{
			DriverId:          driverID,
			DriverPhone:       req.DriverPhone,
			DriverMail:        req.DriverMail,
			VehicleId:         vehicleID,
			CurrentLocation:   req.CurrentLocation.toPb(),
			Destination:       req.Destination.toPb(),
			AvailableVolumeM3: req.AvailableVolumeM3,
			AvailableWeightKg: req.AvailableWeightKg,
			MinPrice:          req.MinPrice,
			DesiredDeposit:    req.DesiredDeposit,
		},
	})
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.OK(ctx, gin.H{
		"ask_id": bytesToUUIDString(resp.AskId),
		"status": resp.Status,
	})
}

type SubmitOfferReq struct {
	BidID        string  `json:"bid_id" binding:"required"`
	AskID        string  `json:"ask_id" binding:"required"`
	DesiredPrice float64 `json:"desired_price" binding:"required,gt=0"`
}

// SubmitOffer godoc
// @Summary      Tài xế báo giá cho một đơn hàng
// @Tags         Matching
// @Accept       json
// @Produce      json
// @Success      200 {object} response.Envelope
// @Router       /api/v1/matching/offers [post]
func (c *MatchingController) SubmitOffer(ctx *gin.Context) {
	var req SubmitOfferReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "VALIDATION_FAILED", err.Error())
		return
	}

	bidID, ok1 := uuidBytes(req.BidID)
	askID, ok2 := uuidBytes(req.AskID)
	if !ok1 || !ok2 {
		response.BadRequest(ctx, "INVALID_ID", "bid_id và ask_id phải là UUID hợp lệ")
		return
	}

	resp, err := c.matchingClient.SubmitOffer(ctx.Request.Context(), &pb.SubmitOfferRequest{
		BidId:        bidID,
		AskId:        askID,
		DesiredPrice: req.DesiredPrice,
	})
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OKMessage(ctx, gin.H{"status": resp.Status}, resp.Message)
}

type RejectOfferReq struct {
	BidID  string `json:"bid_id" binding:"required"`
	AskID  string `json:"ask_id" binding:"required"`
	Reason string `json:"reason"`
}

// RejectOffer godoc
// @Summary      Chủ hàng từ chối báo giá
// @Tags         Matching
// @Accept       json
// @Produce      json
// @Success      200 {object} response.Envelope
// @Router       /api/v1/matching/offers/reject [post]
func (c *MatchingController) RejectOffer(ctx *gin.Context) {
	var req RejectOfferReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "VALIDATION_FAILED", err.Error())
		return
	}

	bidID, ok1 := uuidBytes(req.BidID)
	askID, ok2 := uuidBytes(req.AskID)
	if !ok1 || !ok2 {
		response.BadRequest(ctx, "INVALID_ID", "bid_id và ask_id phải là UUID hợp lệ")
		return
	}

	resp, err := c.matchingClient.RejectOffer(ctx.Request.Context(), &pb.RejectOfferRequest{
		BidId:  bidID,
		AskId:  askID,
		Reason: req.Reason,
	})
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OKMessage(ctx, gin.H{"status": resp.Status}, resp.Message)
}

type AcceptMatchReq struct {
	BidID            string  `json:"bid_id" binding:"required"`
	AskID            string  `json:"ask_id" binding:"required"`
	ConsensusPrice   float64 `json:"consensus_price"`
	ConsensusDeposit float64 `json:"consensus_deposit"`
	ShipperSignature string  `json:"shipper_signature"`
}

// AcceptMatch godoc
// @Summary      Chủ hàng chốt xe
// @Description  Chốt xong, matching_service phát matching.match.found; notification_service báo cho cả chủ hàng lẫn tài xế.
// @Tags         Matching
// @Accept       json
// @Produce      json
// @Success      200 {object} response.Envelope
// @Router       /api/v1/matching/matches/accept [post]
func (c *MatchingController) AcceptMatch(ctx *gin.Context) {
	var req AcceptMatchReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "VALIDATION_FAILED", err.Error())
		return
	}

	bidID, ok1 := uuidBytes(req.BidID)
	askID, ok2 := uuidBytes(req.AskID)
	if !ok1 || !ok2 {
		response.BadRequest(ctx, "INVALID_ID", "bid_id và ask_id phải là UUID hợp lệ")
		return
	}
	shipperID, _ := uuidBytes(middleware.CurrentUserID(ctx))

	resp, err := c.matchingClient.AcceptMatch(ctx.Request.Context(), &pb.AcceptMatchRequest{
		BidId:            bidID,
		AskId:            askID,
		ShipperId:        shipperID,
		ConsensusPrice:   req.ConsensusPrice,
		ConsensusDeposit: req.ConsensusDeposit,
		ShipperSignature: req.ShipperSignature,
		AgreedAt:         timestamppb.Now(),
	})
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.OK(ctx, gin.H{
		"contract_id":       bytesToUUIDString(resp.ContractId),
		"bid_id":            bytesToUUIDString(resp.BidId),
		"ask_id":            bytesToUUIDString(resp.AskId),
		"consensus_price":   resp.ConsensusPrice,
		"consensus_deposit": resp.ConsensusDeposit,
	})
}

func bytesToUUIDString(b []byte) string {
	if len(b) != 16 {
		return ""
	}
	id, err := uuid.FromBytes(b)
	if err != nil {
		return ""
	}
	return id.String()
}