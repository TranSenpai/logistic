package controller

import (
	"strconv"

	"gateway_service/internal/middleware"
	"gateway_service/internal/response"

	"github.com/gin-gonic/gin"
	pb "github.com/logistic/api/logistic/user_service/v1"
)

// UserController là lớp mỏng: bind JSON -> gọi gRPC -> trả JSON.
//
// Không có `if err != nil { ctx.JSON(500, ...) }` nào ở đây nữa: response.Error
// đọc gRPC status + ErrorInfo do user_service gắn và tự chọn đúng HTTP status.
type UserController struct {
	userClient pb.UserServiceClient
}

func NewUserController(userClient pb.UserServiceClient) *UserController {
	return &UserController{userClient: userClient}
}

// queryInt đọc tham số phân trang từ query string, bỏ qua giá trị rác.
func queryInt(ctx *gin.Context, key string) int32 {
	raw := ctx.Query(key)
	if raw == "" {
		return 0
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return 0
	}
	return int32(n)
}

func queryBool(ctx *gin.Context, key string) bool {
	v, _ := strconv.ParseBool(ctx.Query(key))
	return v
}

// resolveUserID lấy user_id theo thứ tự: tham số đường dẫn -> query -> danh tính
// của người đang đăng nhập. Nhờ vậy app gọi /me không cần tự truyền id.
func resolveUserID(ctx *gin.Context, pathKey string) string {
	if v := ctx.Param(pathKey); v != "" {
		return v
	}
	// Route đăng ký tham số là :user_id, nhưng một số handler dùng chung được gọi
	// từ đường dẫn có :id. Thử cả hai để controller không phụ thuộc tên tham số.
	for _, key := range []string{"user_id", "id"} {
		if v := ctx.Param(key); v != "" {
			return v
		}
	}
	if v := ctx.Query("user_id"); v != "" {
		return v
	}
	return middleware.CurrentUserID(ctx)
}

// ===========================================================================
// CLIENT
// ===========================================================================

type RegisterUserReq struct {
	Email    string `json:"email" binding:"omitempty,email"`
	Password string `json:"password" binding:"required,min=6"`
	Phone    string `json:"phone" binding:"required"`
	Role     string `json:"role" binding:"required,oneof=driver shipper"`
	FullName string `json:"full_name"`
}

// RegisterUser godoc
// @Summary      Đăng ký người dùng
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        request body RegisterUserReq true "Thông tin đăng ký"
// @Success      201 {object} response.Envelope
// @Router       /api/v1/users/register [post]
func (c *UserController) RegisterUser(ctx *gin.Context) {
	var req RegisterUserReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "VALIDATION_FAILED", err.Error())
		return
	}

	resp, err := c.userClient.RegisterUser(ctx.Request.Context(), &pb.RegisterUserRequest{
		Email:    req.Email,
		Password: req.Password,
		Phone:    req.Phone,
		Role:     req.Role,
		FullName: req.FullName,
	})
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Created(ctx, gin.H{"id": resp.Id, "user": resp.User}, resp.Message)
}

// GetUser godoc
// @Summary      Lấy thông tin người dùng
// @Tags         User
// @Produce      json
// @Param        id path string true "User ID"
// @Success      200 {object} response.Envelope
// @Router       /api/v1/users/{id} [get]
func (c *UserController) GetUser(ctx *gin.Context) {
	resp, err := c.userClient.GetUser(ctx.Request.Context(), &pb.GetUserRequest{
		Id: resolveUserID(ctx, "id"),
	})
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.OK(ctx, gin.H{
		"user":            resp.User,
		"driver_profile":  resp.DriverProfile,
		"shipper_profile": resp.ShipperProfile,
	})
}

type UpdateUserReq struct {
	FullName  string `json:"full_name"`
	Email     string `json:"email" binding:"omitempty,email"`
	AvatarURL string `json:"avatar_url"`
}

// UpdateUser godoc
// @Summary      Cập nhật thông tin người dùng
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        id path string true "User ID"
// @Success      200 {object} response.Envelope
// @Router       /api/v1/users/{id} [put]
func (c *UserController) UpdateUser(ctx *gin.Context) {
	var req UpdateUserReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "VALIDATION_FAILED", err.Error())
		return
	}

	resp, err := c.userClient.UpdateUser(ctx.Request.Context(), &pb.UpdateUserRequest{
		Id:        resolveUserID(ctx, "id"),
		FullName:  req.FullName,
		Email:     req.Email,
		AvatarUrl: req.AvatarURL,
	})
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.OKMessage(ctx, resp.User, resp.Message)
}

// GetDriverProfile godoc
// @Summary      Lấy hồ sơ tài xế
// @Tags         User
// @Produce      json
// @Param        user_id path string true "User ID"
// @Success      200 {object} response.Envelope
// @Router       /api/v1/users/{user_id}/driver-profile [get]
func (c *UserController) GetDriverProfile(ctx *gin.Context) {
	resp, err := c.userClient.GetDriverProfile(ctx.Request.Context(), &pb.GetDriverProfileRequest{
		UserId: resolveUserID(ctx, "user_id"),
	})
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, resp.DriverProfile)
}

type UpdateDriverProfileReq struct {
	LicenseNumber string `json:"license_number"`
	IDCard        string `json:"id_card"`
}

// UpdateDriverProfile godoc
// @Summary      Cập nhật hồ sơ tài xế
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        user_id path string true "User ID"
// @Success      200 {object} response.Envelope
// @Router       /api/v1/users/{user_id}/driver-profile [put]
func (c *UserController) UpdateDriverProfile(ctx *gin.Context) {
	var req UpdateDriverProfileReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "VALIDATION_FAILED", err.Error())
		return
	}

	resp, err := c.userClient.UpdateDriverProfile(ctx.Request.Context(), &pb.UpdateDriverProfileRequest{
		UserId:        resolveUserID(ctx, "user_id"),
		LicenseNumber: req.LicenseNumber,
		IdCard:        req.IDCard,
	})
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OKMessage(ctx, resp.DriverProfile, resp.Message)
}

// GetShipperProfile godoc
// @Summary      Lấy hồ sơ chủ hàng
// @Tags         User
// @Produce      json
// @Param        user_id path string true "User ID"
// @Success      200 {object} response.Envelope
// @Router       /api/v1/users/{user_id}/shipper-profile [get]
func (c *UserController) GetShipperProfile(ctx *gin.Context) {
	resp, err := c.userClient.GetShipperProfile(ctx.Request.Context(), &pb.GetShipperProfileRequest{
		UserId: resolveUserID(ctx, "user_id"),
	})
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, resp.ShipperProfile)
}

type UpdateShipperProfileReq struct {
	CompanyName     string `json:"company_name"`
	TaxCode         string `json:"tax_code"`
	BusinessAddress string `json:"business_address"`
}

// UpdateShipperProfile godoc
// @Summary      Cập nhật hồ sơ chủ hàng
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        user_id path string true "User ID"
// @Success      200 {object} response.Envelope
// @Router       /api/v1/users/{user_id}/shipper-profile [put]
func (c *UserController) UpdateShipperProfile(ctx *gin.Context) {
	var req UpdateShipperProfileReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "VALIDATION_FAILED", err.Error())
		return
	}

	resp, err := c.userClient.UpdateShipperProfile(ctx.Request.Context(), &pb.UpdateShipperProfileRequest{
		UserId:          resolveUserID(ctx, "user_id"),
		CompanyName:     req.CompanyName,
		TaxCode:         req.TaxCode,
		BusinessAddress: req.BusinessAddress,
	})
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OKMessage(ctx, resp.ShipperProfile, resp.Message)
}

type UpdateDriverKYCReq struct {
	KycStatus string `json:"kyc_status" binding:"required,oneof=pending approved rejected"`
	Note      string `json:"note"`
}

// UpdateDriverKYC godoc
// @Summary      Cập nhật trạng thái KYC
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        user_id path string true "User ID"
// @Success      200 {object} response.Envelope
// @Router       /api/v1/users/{user_id}/kyc [put]
func (c *UserController) UpdateDriverKYC(ctx *gin.Context) {
	var req UpdateDriverKYCReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "VALIDATION_FAILED", err.Error())
		return
	}

	resp, err := c.userClient.UpdateDriverKYC(ctx.Request.Context(), &pb.UpdateDriverKYCRequest{
		UserId:    resolveUserID(ctx, "user_id"),
		KycStatus: req.KycStatus,
		Note:      req.Note,
	})
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OKMessage(ctx, resp.DriverProfile, resp.Message)
}

type AddressReq struct {
	Label        string  `json:"label"`
	ContactName  string  `json:"contact_name"`
	ContactPhone string  `json:"contact_phone"`
	Line1        string  `json:"line1" binding:"required"`
	Ward         string  `json:"ward"`
	District     string  `json:"district"`
	City         string  `json:"city"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	AddressType  string  `json:"address_type" binding:"omitempty,oneof=pickup delivery both"`
	IsDefault    bool    `json:"is_default"`
}

// CreateAddress godoc
// @Summary      Thêm địa chỉ vào sổ địa chỉ
// @Tags         Address
// @Accept       json
// @Produce      json
// @Param        user_id path string true "User ID"
// @Success      201 {object} response.Envelope
// @Router       /api/v1/users/{user_id}/addresses [post]
func (c *UserController) CreateAddress(ctx *gin.Context) {
	var req AddressReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "VALIDATION_FAILED", err.Error())
		return
	}

	resp, err := c.userClient.CreateAddress(ctx.Request.Context(), &pb.CreateAddressRequest{
		UserId:       resolveUserID(ctx, "user_id"),
		Label:        req.Label,
		ContactName:  req.ContactName,
		ContactPhone: req.ContactPhone,
		Line1:        req.Line1,
		Ward:         req.Ward,
		District:     req.District,
		City:         req.City,
		Latitude:     req.Latitude,
		Longitude:    req.Longitude,
		AddressType:  req.AddressType,
		IsDefault:    req.IsDefault,
	})
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Created(ctx, resp.Address, resp.Message)
}

// ListAddresses godoc
// @Summary      Danh sách địa chỉ
// @Tags         Address
// @Produce      json
// @Param        user_id path string true "User ID"
// @Success      200 {object} response.Envelope
// @Router       /api/v1/users/{user_id}/addresses [get]
func (c *UserController) ListAddresses(ctx *gin.Context) {
	resp, err := c.userClient.ListAddresses(ctx.Request.Context(), &pb.ListAddressesRequest{
		UserId:      resolveUserID(ctx, "user_id"),
		AddressType: ctx.Query("address_type"),
		Page:        queryInt(ctx, "page"),
		PageSize:    queryInt(ctx, "page_size"),
	})
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, gin.H{"addresses": resp.Addresses, "pagination": resp.Pagination})
}

// UpdateAddress godoc
// @Summary      Cập nhật địa chỉ
// @Tags         Address
// @Accept       json
// @Produce      json
// @Param        id path string true "Address ID"
// @Success      200 {object} response.Envelope
// @Router       /api/v1/addresses/{id} [put]
func (c *UserController) UpdateAddress(ctx *gin.Context) {
	var req AddressReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "VALIDATION_FAILED", err.Error())
		return
	}

	resp, err := c.userClient.UpdateAddress(ctx.Request.Context(), &pb.UpdateAddressRequest{
		Id:           ctx.Param("id"),
		UserId:       middleware.CurrentUserID(ctx),
		Label:        req.Label,
		ContactName:  req.ContactName,
		ContactPhone: req.ContactPhone,
		Line1:        req.Line1,
		Ward:         req.Ward,
		District:     req.District,
		City:         req.City,
		Latitude:     req.Latitude,
		Longitude:    req.Longitude,
		AddressType:  req.AddressType,
		IsDefault:    req.IsDefault,
	})
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OKMessage(ctx, resp.Address, resp.Message)
}

// DeleteAddress godoc
// @Summary      Xoá địa chỉ
// @Tags         Address
// @Produce      json
// @Param        id path string true "Address ID"
// @Success      200 {object} response.Envelope
// @Router       /api/v1/addresses/{id} [delete]
func (c *UserController) DeleteAddress(ctx *gin.Context) {
	resp, err := c.userClient.DeleteAddress(ctx.Request.Context(), &pb.DeleteAddressRequest{
		Id:     ctx.Param("id"),
		UserId: middleware.CurrentUserID(ctx),
	})
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OKMessage(ctx, nil, resp.Message)
}

type RegisterDeviceReq struct {
	DeviceToken string `json:"device_token" binding:"required"`
	Platform    string `json:"platform" binding:"omitempty,oneof=android ios web"`
	DeviceName  string `json:"device_name"`
}

// RegisterDevice godoc
// @Summary      Đăng ký thiết bị nhận push
// @Tags         Device
// @Accept       json
// @Produce      json
// @Param        user_id path string true "User ID"
// @Success      201 {object} response.Envelope
// @Router       /api/v1/users/{user_id}/devices [post]
func (c *UserController) RegisterDevice(ctx *gin.Context) {
	var req RegisterDeviceReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "VALIDATION_FAILED", err.Error())
		return
	}

	resp, err := c.userClient.RegisterDevice(ctx.Request.Context(), &pb.RegisterDeviceRequest{
		UserId:      resolveUserID(ctx, "user_id"),
		DeviceToken: req.DeviceToken,
		Platform:    req.Platform,
		DeviceName:  req.DeviceName,
	})
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Created(ctx, resp.Device, resp.Message)
}

// ListDevices godoc
// @Summary      Danh sách thiết bị
// @Tags         Device
// @Produce      json
// @Param        user_id path string true "User ID"
// @Success      200 {object} response.Envelope
// @Router       /api/v1/users/{user_id}/devices [get]
func (c *UserController) ListDevices(ctx *gin.Context) {
	resp, err := c.userClient.ListDevices(ctx.Request.Context(), &pb.ListDevicesRequest{
		UserId: resolveUserID(ctx, "user_id"),
	})
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, gin.H{"devices": resp.Devices})
}

// DeleteDevice godoc
// @Summary      Xoá thiết bị
// @Tags         Device
// @Produce      json
// @Param        id path string true "Device ID"
// @Success      200 {object} response.Envelope
// @Router       /api/v1/devices/{id} [delete]
func (c *UserController) DeleteDevice(ctx *gin.Context) {
	resp, err := c.userClient.DeleteDevice(ctx.Request.Context(), &pb.DeleteDeviceRequest{
		Id:     ctx.Param("id"),
		UserId: middleware.CurrentUserID(ctx),
	})
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OKMessage(ctx, nil, resp.Message)
}

// ===========================================================================
// ADMIN
// ===========================================================================

// AdminListUsers godoc
// @Summary      [Admin] Danh sách người dùng
// @Tags         Admin-User
// @Produce      json
// @Success      200 {object} response.Envelope
// @Router       /api/v1/admin/users [get]
func (c *UserController) AdminListUsers(ctx *gin.Context) {
	resp, err := c.userClient.AdminListUsers(ctx.Request.Context(), &pb.AdminListUsersRequest{
		Role:     ctx.Query("role"),
		Status:   ctx.Query("status"),
		Keyword:  ctx.Query("keyword"),
		Page:     queryInt(ctx, "page"),
		PageSize: queryInt(ctx, "page_size"),
	})
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, gin.H{"users": resp.Users, "pagination": resp.Pagination})
}

type AdminUpdateUserStatusReq struct {
	Status string `json:"status" binding:"required,oneof=active banned suspended"`
	Reason string `json:"reason"`
}

// AdminUpdateUserStatus godoc
// @Summary      [Admin] Khoá/mở tài khoản
// @Tags         Admin-User
// @Accept       json
// @Produce      json
// @Param        id path string true "User ID"
// @Success      200 {object} response.Envelope
// @Router       /api/v1/admin/users/{id}/status [put]
func (c *UserController) AdminUpdateUserStatus(ctx *gin.Context) {
	var req AdminUpdateUserStatusReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "VALIDATION_FAILED", err.Error())
		return
	}

	resp, err := c.userClient.AdminUpdateUserStatus(ctx.Request.Context(), &pb.AdminUpdateUserStatusRequest{
		Id:     ctx.Param("id"),
		Status: req.Status,
		Reason: req.Reason,
	})
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OKMessage(ctx, resp.User, resp.Message)
}

// AdminListPendingKYC godoc
// @Summary      [Admin] Hàng đợi duyệt KYC
// @Tags         Admin-User
// @Produce      json
// @Success      200 {object} response.Envelope
// @Router       /api/v1/admin/kyc/pending [get]
func (c *UserController) AdminListPendingKYC(ctx *gin.Context) {
	resp, err := c.userClient.AdminListPendingKYC(ctx.Request.Context(), &pb.AdminListPendingKYCRequest{
		Page:     queryInt(ctx, "page"),
		PageSize: queryInt(ctx, "page_size"),
	})
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, gin.H{"driver_profiles": resp.DriverProfiles, "pagination": resp.Pagination})
}

type AdminReviewKYCReq struct {
	Approved bool   `json:"approved"`
	Note     string `json:"note"`
}

// AdminReviewKYC godoc
// @Summary      [Admin] Duyệt/từ chối KYC
// @Tags         Admin-User
// @Accept       json
// @Produce      json
// @Param        user_id path string true "User ID"
// @Success      200 {object} response.Envelope
// @Router       /api/v1/admin/kyc/{user_id}/review [put]
func (c *UserController) AdminReviewKYC(ctx *gin.Context) {
	var req AdminReviewKYCReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "VALIDATION_FAILED", err.Error())
		return
	}

	resp, err := c.userClient.AdminReviewKYC(ctx.Request.Context(), &pb.AdminReviewKYCRequest{
		UserId:     ctx.Param("user_id"),
		Approved:   req.Approved,
		Note:       req.Note,
		ReviewerId: middleware.CurrentUserID(ctx),
	})
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OKMessage(ctx, resp.DriverProfile, resp.Message)
}

// AdminGetUserStats godoc
// @Summary      [Admin] Thống kê người dùng
// @Tags         Admin-User
// @Produce      json
// @Success      200 {object} response.Envelope
// @Router       /api/v1/admin/users/stats [get]
func (c *UserController) AdminGetUserStats(ctx *gin.Context) {
	resp, err := c.userClient.AdminGetUserStats(ctx.Request.Context(), &pb.AdminGetUserStatsRequest{})
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, resp)
}

// AdminDeleteUser godoc
// @Summary      [Admin] Xoá người dùng
// @Tags         Admin-User
// @Produce      json
// @Param        id path string true "User ID"
// @Success      200 {object} response.Envelope
// @Router       /api/v1/admin/users/{id} [delete]
func (c *UserController) AdminDeleteUser(ctx *gin.Context) {
	resp, err := c.userClient.AdminDeleteUser(ctx.Request.Context(), &pb.AdminDeleteUserRequest{
		Id: ctx.Param("id"),
	})
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OKMessage(ctx, nil, resp.Message)
}

var _ = queryBool
