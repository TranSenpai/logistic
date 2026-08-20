// Package controller là lớp vỏ gRPC của vehicle_service.
// Không có luật nghiệp vụ, không có xử lý lỗi transport — xem ghi chú ở
// user_service/internal/controller/user_controller.go.
package controller

import (
	"context"

	"vehicle_service/internal/biz"
	cerr "vehicle_service/internal/common/errors"
	"vehicle_service/internal/mapper"

	"github.com/google/uuid"
	pb "github.com/logistic/api/logistic/vehicle_service/v1"
)

type vehicleController struct {
	pb.UnimplementedVehicleServiceServer
	engine biz.VehicleEngine
	mapper mapper.AppMapper
}

func NewVehicleController(engine biz.VehicleEngine, appMapper mapper.AppMapper) pb.VehicleServiceServer {
	return &vehicleController{engine: engine, mapper: appMapper}
}

func parseID(raw string, invalid error) (uuid.UUID, error) {
	if raw == "" {
		return uuid.Nil, invalid
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, invalid
	}
	return id, nil
}

// parseOptionalID: id rỗng là hợp lệ (nghĩa là "bỏ qua kiểm tra sở hữu"),
// nhưng id có mà sai định dạng thì vẫn phải báo lỗi.
func parseOptionalID(raw string, invalid error) (uuid.UUID, error) {
	if raw == "" {
		return uuid.Nil, nil
	}
	return parseID(raw, invalid)
}

// ===========================================================================
// CLIENT — PHƯƠNG TIỆN
// ===========================================================================

func (c *vehicleController) RegisterVehicle(ctx context.Context, req *pb.RegisterVehicleRequest) (*pb.RegisterVehicleResponse, error) {
	param, err := c.mapper.PbRegisterVehicleToParam(req)
	if err != nil {
		return nil, cerr.ErrInvalidDriverID.WithCause(err)
	}

	v, err := c.engine.RegisterVehicle(ctx, &param)
	if err != nil {
		return nil, err
	}

	return &pb.RegisterVehicleResponse{
		Id:      v.ID.String(),
		Message: "Đăng ký phương tiện thành công",
		Vehicle: c.mapper.EntityVehicleToPbVehicle(*v),
	}, nil
}

func (c *vehicleController) GetVehicle(ctx context.Context, req *pb.GetVehicleRequest) (*pb.GetVehicleResponse, error) {
	id, err := parseID(req.Id, cerr.ErrInvalidVehicleID)
	if err != nil {
		return nil, err
	}

	v, err := c.engine.GetVehicle(ctx, id)
	if err != nil {
		return nil, err
	}
	return &pb.GetVehicleResponse{Vehicle: c.mapper.EntityVehicleToPbVehicle(*v)}, nil
}

func (c *vehicleController) ListVehicles(ctx context.Context, req *pb.ListVehiclesRequest) (*pb.ListVehiclesResponse, error) {
	param, err := c.mapper.PbListVehiclesToParam(req)
	if err != nil {
		return nil, cerr.ErrInvalidDriverID.WithCause(err)
	}

	res, err := c.engine.ListVehicles(ctx, &param)
	if err != nil {
		return nil, err
	}

	return &pb.ListVehiclesResponse{
		Vehicles:   c.mapper.EntityVehicleListToPbVehicleList(res.Vehicles),
		Pagination: c.mapper.EntityPaginationToPb(res.Pagination),
	}, nil
}

func (c *vehicleController) UpdateVehicle(ctx context.Context, req *pb.UpdateVehicleRequest) (*pb.UpdateVehicleResponse, error) {
	param, err := c.mapper.PbUpdateVehicleToParam(req)
	if err != nil {
		return nil, cerr.ErrInvalidVehicleID.WithCause(err)
	}

	v, err := c.engine.UpdateVehicle(ctx, &param)
	if err != nil {
		return nil, err
	}

	return &pb.UpdateVehicleResponse{
		Vehicle: c.mapper.EntityVehicleToPbVehicle(*v),
		Message: "Cập nhật phương tiện thành công",
	}, nil
}

func (c *vehicleController) DeleteVehicle(ctx context.Context, req *pb.DeleteVehicleRequest) (*pb.DeleteVehicleResponse, error) {
	id, err := parseID(req.Id, cerr.ErrInvalidVehicleID)
	if err != nil {
		return nil, err
	}
	driverID, err := parseOptionalID(req.DriverId, cerr.ErrInvalidDriverID)
	if err != nil {
		return nil, err
	}

	if err := c.engine.DeleteVehicle(ctx, id, driverID); err != nil {
		return nil, err
	}
	return &pb.DeleteVehicleResponse{Message: "Xoá phương tiện thành công"}, nil
}

func (c *vehicleController) UpdateVehicleStatus(ctx context.Context, req *pb.UpdateVehicleStatusRequest) (*pb.UpdateVehicleStatusResponse, error) {
	id, err := parseID(req.Id, cerr.ErrInvalidVehicleID)
	if err != nil {
		return nil, err
	}

	v, err := c.engine.UpdateVehicleStatus(ctx, id, req.Status)
	if err != nil {
		return nil, err
	}

	return &pb.UpdateVehicleStatusResponse{
		Message: "Cập nhật trạng thái phương tiện thành công",
		Vehicle: c.mapper.EntityVehicleToPbVehicle(*v),
	}, nil
}

// ===========================================================================
// CLIENT — GIẤY TỜ
// ===========================================================================

func (c *vehicleController) UploadVehicleDocument(ctx context.Context, req *pb.UploadVehicleDocumentRequest) (*pb.UploadVehicleDocumentResponse, error) {
	param, err := c.mapper.PbUploadDocumentToParam(req)
	if err != nil {
		return nil, cerr.ErrInvalidVehicleID.WithCause(err)
	}

	doc, err := c.engine.UploadDocument(ctx, &param)
	if err != nil {
		return nil, err
	}

	return &pb.UploadVehicleDocumentResponse{
		Document: c.mapper.EntityDocumentToPb(*doc),
		Message:  "Tải lên giấy tờ thành công, đang chờ duyệt",
	}, nil
}

func (c *vehicleController) ListVehicleDocuments(ctx context.Context, req *pb.ListVehicleDocumentsRequest) (*pb.ListVehicleDocumentsResponse, error) {
	param, err := c.mapper.PbListDocumentsToParam(req)
	if err != nil {
		return nil, cerr.ErrInvalidVehicleID.WithCause(err)
	}

	docs, err := c.engine.ListDocuments(ctx, &param)
	if err != nil {
		return nil, err
	}
	return &pb.ListVehicleDocumentsResponse{Documents: c.mapper.EntityDocumentListToPbList(docs)}, nil
}

func (c *vehicleController) DeleteVehicleDocument(ctx context.Context, req *pb.DeleteVehicleDocumentRequest) (*pb.DeleteVehicleDocumentResponse, error) {
	id, err := parseID(req.Id, cerr.ErrInvalidDocumentID)
	if err != nil {
		return nil, err
	}
	if err := c.engine.DeleteDocument(ctx, id); err != nil {
		return nil, err
	}
	return &pb.DeleteVehicleDocumentResponse{Message: "Xoá giấy tờ thành công"}, nil
}

// ===========================================================================
// CLIENT — VỊ TRÍ & SẴN SÀNG NHẬN ĐƠN
// ===========================================================================

func (c *vehicleController) ReportLocation(ctx context.Context, req *pb.ReportLocationRequest) (*pb.ReportLocationResponse, error) {
	param, err := c.mapper.PbReportLocationToParam(req)
	if err != nil {
		return nil, cerr.ErrInvalidVehicleID.WithCause(err)
	}

	loc, err := c.engine.ReportLocation(ctx, &param)
	if err != nil {
		return nil, err
	}

	return &pb.ReportLocationResponse{
		Message: "Cập nhật vị trí thành công",
		ZoneId:  loc.ZoneID,
	}, nil
}

func (c *vehicleController) GetVehicleLocation(ctx context.Context, req *pb.GetVehicleLocationRequest) (*pb.GetVehicleLocationResponse, error) {
	id, err := parseID(req.VehicleId, cerr.ErrInvalidVehicleID)
	if err != nil {
		return nil, err
	}

	loc, err := c.engine.GetLocation(ctx, id)
	if err != nil {
		return nil, err
	}
	return &pb.GetVehicleLocationResponse{Location: c.mapper.EntityLocationToPb(*loc)}, nil
}

func (c *vehicleController) SetDriverAvailability(ctx context.Context, req *pb.SetDriverAvailabilityRequest) (*pb.SetDriverAvailabilityResponse, error) {
	param, err := c.mapper.PbSetAvailabilityToParam(req)
	if err != nil {
		return nil, cerr.ErrInvalidDriverID.WithCause(err)
	}

	avail, err := c.engine.SetAvailability(ctx, &param)
	if err != nil {
		return nil, err
	}

	message := "Đã tắt nhận đơn"
	if avail.IsOnline {
		message = "Đã bật nhận đơn, xe của bạn sẽ xuất hiện trong kết quả tìm kiếm"
	}

	return &pb.SetDriverAvailabilityResponse{
		Availability: c.mapper.EntityAvailabilityToPb(*avail),
		Message:      message,
	}, nil
}

func (c *vehicleController) GetDriverAvailability(ctx context.Context, req *pb.GetDriverAvailabilityRequest) (*pb.GetDriverAvailabilityResponse, error) {
	id, err := parseID(req.DriverId, cerr.ErrInvalidDriverID)
	if err != nil {
		return nil, err
	}

	avail, err := c.engine.GetAvailability(ctx, id)
	if err != nil {
		return nil, err
	}
	return &pb.GetDriverAvailabilityResponse{Availability: c.mapper.EntityAvailabilityToPb(*avail)}, nil
}

// SearchNearbyVehicles là endpoint matching_service gọi sang khi có đơn hàng mới.
func (c *vehicleController) SearchNearbyVehicles(ctx context.Context, req *pb.SearchNearbyVehiclesRequest) (*pb.SearchNearbyVehiclesResponse, error) {
	param := c.mapper.PbSearchNearbyToParam(req)

	list, err := c.engine.SearchNearby(ctx, &param)
	if err != nil {
		return nil, err
	}

	return &pb.SearchNearbyVehiclesResponse{
		Vehicles:   c.mapper.EntityNearbyListToPbList(list),
		TotalFound: int32(len(list)),
	}, nil
}

// ===========================================================================
// ADMIN
// ===========================================================================

func (c *vehicleController) AdminListVehicles(ctx context.Context, req *pb.AdminListVehiclesRequest) (*pb.AdminListVehiclesResponse, error) {
	param := c.mapper.PbAdminListVehiclesToParam(req)

	res, err := c.engine.AdminListVehicles(ctx, &param)
	if err != nil {
		return nil, err
	}

	return &pb.AdminListVehiclesResponse{
		Vehicles:   c.mapper.EntityVehicleListToPbVehicleList(res.Vehicles),
		Pagination: c.mapper.EntityPaginationToPb(res.Pagination),
	}, nil
}

func (c *vehicleController) AdminVerifyVehicle(ctx context.Context, req *pb.AdminVerifyVehicleRequest) (*pb.AdminVerifyVehicleResponse, error) {
	param, err := c.mapper.PbVerifyVehicleToParam(req)
	if err != nil {
		return nil, cerr.ErrInvalidVehicleID.WithCause(err)
	}
	// ReviewerID được mapper bỏ qua (goverter:ignore) vì cần parse riêng —
	// admin id có thể trống khi thao tác từ script nội bộ.
	if param.ReviewerID, err = parseOptionalID(req.ReviewerId, cerr.ErrInvalidDriverID); err != nil {
		return nil, err
	}

	v, err := c.engine.AdminVerifyVehicle(ctx, &param)
	if err != nil {
		return nil, err
	}

	message := "Đã từ chối duyệt phương tiện"
	if req.Approved {
		message = "Đã duyệt phương tiện"
	}

	return &pb.AdminVerifyVehicleResponse{
		Vehicle: c.mapper.EntityVehicleToPbVehicle(*v),
		Message: message,
	}, nil
}

func (c *vehicleController) AdminListPendingDocuments(ctx context.Context, req *pb.AdminListPendingDocumentsRequest) (*pb.AdminListPendingDocumentsResponse, error) {
	res, err := c.engine.AdminListPendingDocuments(ctx, int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, err
	}

	return &pb.AdminListPendingDocumentsResponse{
		Documents:  c.mapper.EntityDocumentListToPbList(res.Documents),
		Pagination: c.mapper.EntityPaginationToPb(res.Pagination),
	}, nil
}

func (c *vehicleController) AdminReviewDocument(ctx context.Context, req *pb.AdminReviewDocumentRequest) (*pb.AdminReviewDocumentResponse, error) {
	param, err := c.mapper.PbReviewDocumentToParam(req)
	if err != nil {
		return nil, cerr.ErrInvalidDocumentID.WithCause(err)
	}
	if param.ReviewerID, err = parseOptionalID(req.ReviewerId, cerr.ErrInvalidDriverID); err != nil {
		return nil, err
	}

	doc, err := c.engine.AdminReviewDocument(ctx, &param)
	if err != nil {
		return nil, err
	}

	message := "Đã từ chối giấy tờ"
	if req.Approved {
		message = "Đã duyệt giấy tờ"
	}

	return &pb.AdminReviewDocumentResponse{
		Document: c.mapper.EntityDocumentToPb(*doc),
		Message:  message,
	}, nil
}

func (c *vehicleController) AdminGetVehicleStats(ctx context.Context, _ *pb.AdminGetVehicleStatsRequest) (*pb.AdminGetVehicleStatsResponse, error) {
	stats, err := c.engine.AdminGetStats(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.AdminGetVehicleStatsResponse{
		TotalVehicles:       stats.TotalVehicles,
		ActiveVehicles:      stats.ActiveVehicles,
		MaintenanceVehicles: stats.MaintenanceVehicles,
		PendingVerification: stats.PendingVerification,
		OnlineDrivers:       stats.OnlineDrivers,
		PendingDocuments:    stats.PendingDocuments,
	}, nil
}
