package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	pb "github.com/logistic/api/logistic/user_service/v1"
	"google.golang.org/grpc/status"
)

type UserController struct {
	userClient pb.UserServiceClient
}

func NewUserController(userClient pb.UserServiceClient) *UserController {
	return &UserController{
		userClient: userClient,
	}
}

// Structs for parsing HTTP requests
type RegisterUserReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Phone    string `json:"phone" binding:"required"`
	Role     string `json:"role" binding:"required"`
}

type UpdateDriverKYCReq struct {
	KycStatus string `json:"kyc_status" binding:"required"`
}

// RegisterUser godoc
// @Summary      Đăng ký user mới
// @Description  Đăng ký người dùng hoặc tài xế mới vào hệ thống.
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        request body RegisterUserReq true "Thông tin đăng ký"
// @Success      201 {object} map[string]interface{} "Tạo người dùng thành công"
// @Failure      400 {object} map[string]interface{} "Lỗi dữ liệu đầu vào"
// @Failure      500 {object} map[string]interface{} "Lỗi server nội bộ"
// @Router       /api/user/v1/register [post]
func (c *UserController) RegisterUser(ctx *gin.Context) {
	var req RegisterUserReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "validation_failed", "message": err.Error()})
		return
	}

	resp, err := c.userClient.RegisterUser(ctx.Request.Context(), &pb.RegisterUserRequest{
		Email:    req.Email,
		Password: req.Password,
		Phone:    req.Phone,
		Role:     req.Role,
	})
	if err != nil {
		st, _ := status.FromError(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "registration_failed", "message": st.Message()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": resp.Message,
		"user_id": resp.Id,
	})
}

// GetUser godoc
// @Summary      Lấy thông tin User
// @Description  Lấy thông tin chi tiết của người dùng bằng ID.
// @Tags         User
// @Produce      json
// @Param        id path string true "User ID"
// @Success      200 {object} map[string]interface{} "Thông tin chi tiết người dùng"
// @Failure      400 {object} map[string]interface{} "Thiếu ID"
// @Failure      500 {object} map[string]interface{} "Lỗi server nội bộ"
// @Router       /api/user/v1/{id} [get]
func (c *UserController) GetUser(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "validation_failed", "message": "id is required"})
		return
	}

	resp, err := c.userClient.GetUser(ctx.Request.Context(), &pb.GetUserRequest{
		Id: id,
	})
	if err != nil {
		st, _ := status.FromError(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "fetch_failed", "message": st.Message()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"user": resp.User,
		"driver_profile": resp.DriverProfile,
		"shipper_profile": resp.ShipperProfile,
	})
}

// UpdateDriverKYC godoc
// @Summary      Cập nhật thông tin KYC của tài xế
// @Description  Cập nhật trạng thái KYC.
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        user_id path string true "User ID của tài xế"
// @Param        request body UpdateDriverKYCReq true "Thông tin KYC"
// @Success      200 {object} map[string]interface{} "Cập nhật thành công"
// @Failure      400 {object} map[string]interface{} "Lỗi dữ liệu đầu vào"
// @Failure      500 {object} map[string]interface{} "Lỗi server nội bộ"
// @Router       /api/user/v1/{user_id}/kyc [put]
func (c *UserController) UpdateDriverKYC(ctx *gin.Context) {
	userID := ctx.Param("user_id")
	if userID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "validation_failed", "message": "user_id is required"})
		return
	}

	var req UpdateDriverKYCReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "validation_failed", "message": err.Error()})
		return
	}

	resp, err := c.userClient.UpdateDriverKYC(ctx.Request.Context(), &pb.UpdateDriverKYCRequest{
		UserId:    userID,
		KycStatus: req.KycStatus,
	})
	if err != nil {
		st, _ := status.FromError(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "update_failed", "message": st.Message()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": resp.Message,
	})
}
