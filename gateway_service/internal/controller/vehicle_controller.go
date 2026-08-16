package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	pb "github.com/logistic/api/logistic/vehicle_service/v1"
	"google.golang.org/grpc/status"
)

type VehicleController struct {
	vehicleClient pb.VehicleServiceClient
}

func NewVehicleController(vehicleClient pb.VehicleServiceClient) *VehicleController {
	return &VehicleController{
		vehicleClient: vehicleClient,
	}
}

type RegisterVehicleReq struct {
	DriverId          string  `json:"driver_id" binding:"required"`
	LicensePlate      string  `json:"license_plate" binding:"required"`
	Brand             string  `json:"brand" binding:"required"`
	Model             string  `json:"model" binding:"required"`
	VehicleType       string  `json:"vehicle_type" binding:"required"`
	CapacityWeightKg  float32 `json:"capacity_weight_kg" binding:"required"`
	CapacityVolumeCbm float32 `json:"capacity_volume_cbm" binding:"required"`
}

type UpdateVehicleStatusReq struct {
	Status string `json:"status" binding:"required"`
}

// RegisterVehicle godoc
// @Summary      Đăng ký phương tiện
// @Description  Đăng ký thông tin phương tiện (xe tải, xe khách...) cho tài xế.
// @Tags         Vehicle
// @Accept       json
// @Produce      json
// @Param        request body RegisterVehicleReq true "Thông tin phương tiện"
// @Success      201 {object} map[string]interface{} "Tạo phương tiện thành công"
// @Failure      400 {object} map[string]interface{} "Lỗi dữ liệu đầu vào"
// @Failure      500 {object} map[string]interface{} "Lỗi server nội bộ"
// @Router       /api/vehicle/v1/register [post]
func (c *VehicleController) RegisterVehicle(ctx *gin.Context) {
	var req RegisterVehicleReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "validation_failed", "message": err.Error()})
		return
	}

	resp, err := c.vehicleClient.RegisterVehicle(ctx.Request.Context(), &pb.RegisterVehicleRequest{
		DriverId:          req.DriverId,
		LicensePlate:      req.LicensePlate,
		Brand:             req.Brand,
		Model:             req.Model,
		VehicleType:       req.VehicleType,
		CapacityWeightKg:  req.CapacityWeightKg,
		CapacityVolumeCbm: req.CapacityVolumeCbm,
	})
	if err != nil {
		st, _ := status.FromError(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "registration_failed", "message": st.Message()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message":    resp.Message,
		"vehicle_id": resp.Id,
	})
}

// GetVehicle godoc
// @Summary      Lấy thông tin phương tiện
// @Description  Lấy thông tin chi tiết của một phương tiện theo ID.
// @Tags         Vehicle
// @Produce      json
// @Param        id path string true "Vehicle ID"
// @Success      200 {object} map[string]interface{} "Thông tin chi tiết phương tiện"
// @Failure      400 {object} map[string]interface{} "Thiếu ID"
// @Failure      500 {object} map[string]interface{} "Lỗi server nội bộ"
// @Router       /api/vehicle/v1/{id} [get]
func (c *VehicleController) GetVehicle(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "validation_failed", "message": "id is required"})
		return
	}

	resp, err := c.vehicleClient.GetVehicle(ctx.Request.Context(), &pb.GetVehicleRequest{
		Id: id,
	})
	if err != nil {
		st, _ := status.FromError(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "fetch_failed", "message": st.Message()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"vehicle": resp.Vehicle,
	})
}

// ListVehicles godoc
// @Summary      Danh sách phương tiện
// @Description  Lấy danh sách các phương tiện trong hệ thống của tài xế.
// @Tags         Vehicle
// @Produce      json
// @Param        driver_id query string false "Driver ID để lọc"
// @Success      200 {object} map[string]interface{} "Danh sách phương tiện"
// @Failure      500 {object} map[string]interface{} "Lỗi server nội bộ"
// @Router       /api/vehicle/v1/list [get]
func (c *VehicleController) ListVehicles(ctx *gin.Context) {
	driverId := ctx.Query("driver_id")
	resp, err := c.vehicleClient.ListVehicles(ctx.Request.Context(), &pb.ListVehiclesRequest{
		DriverId: driverId,
	})
	if err != nil {
		st, _ := status.FromError(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "fetch_failed", "message": st.Message()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"vehicles": resp.Vehicles,
	})
}

// UpdateVehicleStatus godoc
// @Summary      Cập nhật trạng thái
// @Description  Cập nhật trạng thái của phương tiện (Active, Inactive, InTransit...).
// @Tags         Vehicle
// @Accept       json
// @Produce      json
// @Param        id path string true "Vehicle ID"
// @Param        request body UpdateVehicleStatusReq true "Trạng thái mới"
// @Success      200 {object} map[string]interface{} "Cập nhật thành công"
// @Failure      400 {object} map[string]interface{} "Lỗi dữ liệu đầu vào"
// @Failure      500 {object} map[string]interface{} "Lỗi server nội bộ"
// @Router       /api/vehicle/v1/{id}/status [put]
func (c *VehicleController) UpdateVehicleStatus(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "validation_failed", "message": "id is required"})
		return
	}

	var req UpdateVehicleStatusReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "validation_failed", "message": err.Error()})
		return
	}

	resp, err := c.vehicleClient.UpdateVehicleStatus(ctx.Request.Context(), &pb.UpdateVehicleStatusRequest{
		Id:     id,
		Status: req.Status,
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
