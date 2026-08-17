package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	pb "github.com/logistic/api/logistic/matching_service/v1"
	"google.golang.org/grpc/status"
)

type MatchingController struct {
	matchingClient pb.MatchingEngineServiceClient
}

func NewMatchingController(matchingClient pb.MatchingEngineServiceClient) *MatchingController {
	return &MatchingController{
		matchingClient: matchingClient,
	}
}

type SubmitBidReq struct {
	ShipperId string  `json:"shipper_id" binding:"required"`
	MaxPrice  float64 `json:"max_price" binding:"required"`
}

type SubmitAskReq struct {
	DriverId string  `json:"driver_id" binding:"required"`
	MinPrice float64 `json:"min_price" binding:"required"`
}

type AcceptMatchReq struct {
	BidId           string  `json:"bid_id" binding:"required"`
	AskId           string  `json:"ask_id" binding:"required"`
	ConsensusPrice  float64 `json:"consensus_price" binding:"required"`
}

// SubmitBid godoc
// @Summary      Gửi yêu cầu vận chuyển (Bid)
// @Description  Khách hàng gửi yêu cầu vận chuyển hàng hóa kèm theo giá cước đề xuất (Bid).
// @Tags         Matching
// @Accept       json
// @Produce      json
// @Param        request body SubmitBidReq true "Thông vị Bid"
// @Success      200 {object} map[string]interface{} "Gửi Bid thành công"
// @Failure      400 {object} map[string]interface{} "Lỗi dữ liệu đầu vào"
// @Failure      500 {object} map[string]interface{} "Lỗi server nội bộ"
// @Router       /api/matching/v1/bid [post]
func (c *MatchingController) SubmitBid(ctx *gin.Context) {
	var req SubmitBidReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "validation_failed", "message": err.Error()})
		return
	}

	resp, err := c.matchingClient.SubmitBid(ctx.Request.Context(), &pb.SubmitBidRequest{
		Payload: &pb.Bid{
			ShipperId: []byte(req.ShipperId),
			MaxPrice:  req.MaxPrice,
		},
	})
	if err != nil {
		st, _ := status.FromError(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "bid_failed", "message": st.Message()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": resp.Status,
		"bid_id": string(resp.BidId),
	})
}

// SubmitAsk godoc
// @Summary      Gửi báo giá vận chuyển (Ask)
// @Description  Tài xế gửi báo giá có xe trống (Ask) để ghép chuyến.
// @Tags         Matching
// @Accept       json
// @Produce      json
// @Param        request body SubmitAskReq true "Thông tin Ask"
// @Success      200 {object} map[string]interface{} "Gửi Ask thành công"
// @Failure      400 {object} map[string]interface{} "Lỗi dữ liệu đầu vào"
// @Failure      500 {object} map[string]interface{} "Lỗi server nội bộ"
// @Router       /api/matching/v1/ask [post]
func (c *MatchingController) SubmitAsk(ctx *gin.Context) {
	var req SubmitAskReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "validation_failed", "message": err.Error()})
		return
	}

	resp, err := c.matchingClient.SubmitAsk(ctx.Request.Context(), &pb.SubmitAskRequest{
		Payload: &pb.Ask{
			DriverId: []byte(req.DriverId),
			MinPrice: req.MinPrice,
		},
	})
	if err != nil {
		st, _ := status.FromError(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "ask_failed", "message": st.Message()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": resp.Status,
		"ask_id": string(resp.AskId),
	})
}

// AcceptMatch godoc
// @Summary      Chấp nhận ghép chuyến
// @Description  Khách hàng hoặc tài xế chấp nhận lệnh ghép chuyến.
// @Tags         Matching
// @Accept       json
// @Produce      json
// @Param        request body AcceptMatchReq true "Thông tin Match"
// @Success      200 {object} map[string]interface{} "Chấp nhận thành công"
// @Failure      400 {object} map[string]interface{} "Lỗi dữ liệu đầu vào"
// @Failure      500 {object} map[string]interface{} "Lỗi server nội bộ"
// @Router       /api/matching/v1/accept [post]
func (c *MatchingController) AcceptMatch(ctx *gin.Context) {
	var req AcceptMatchReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "validation_failed", "message": err.Error()})
		return
	}

	resp, err := c.matchingClient.AcceptMatch(ctx.Request.Context(), &pb.AcceptMatchRequest{
		BidId:          []byte(req.BidId),
		AskId:          []byte(req.AskId),
		ConsensusPrice: req.ConsensusPrice,
	})
	if err != nil {
		st, _ := status.FromError(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "accept_failed", "message": st.Message()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"contract_id": string(resp.ContractId),
	})
}
