package app

import (
	"context"

	"user_service/internal/entity"

	"github.com/google/uuid"
)

// ĐƯỜNG NỐI (seam) — đọc trước khi sửa file này.
//
// Duyệt hồ sơ tài xế (KYC) là một bounded context lồng bên trong "Party": nó có
// máy trạng thái riêng (pending -> approved/rejected), có actor riêng (admin), và
// ngoài đời nó còn phình ra giấy phép, bảo hiểm, hạn bằng lái, lịch sử duyệt,
// tài liệu đính kèm. Đó là lý do nó được tách thành use case riêng ở đây.
//
// Nhưng nó CHƯA tách thành service riêng, và có lý do:
//   - `kyc_status` là một cột trên chính aggregate DriverProfile. Tách service
//     bây giờ là xẻ đôi một aggregate — thứ mà `BM` ch2 trang in 58 nói rõ là
//     không nên: "one microservice can manage one or more aggregates, but we
//     don't want one aggregate to be managed by more than one microservice."
//   - Khối lượng hiện tại là 3 method. Chưa đủ nuôi một service.
//   - gateway_service đang gọi vào user_service ở rất nhiều chỗ; tách đồng nghĩa
//     đổi proto, đổi gateway, migrate dữ liệu — trong khi nửa sau vòng đời
//     nghiệp vụ (giao hàng, thanh toán) còn chưa tồn tại.
//
// Cách làm ở đây là cái `BM` ch2 trang in 58 gọi là "turtles all the way down":
// giữ service ở mức bounded context thô, chia context lồng bên trong, và khi
// tách ra thì vẫn giấu được quyết định đó sau một API thô hơn.
//
// KHI NÀO THÌ TÁCH THẬT: khi compliance có bảng dữ liệu riêng (không còn là một
// cột trên driver_profiles), hoặc khi nó cần nhịp deploy khác phần hồ sơ. Lúc đó
// ComplianceRepository trong port.go chính là danh sách RPC cần dựng.
type ComplianceUseCase interface {
	UpdateDriverKYC(ctx context.Context, param *entity.UpdateDriverKYCParam) (*entity.DriverProfile, error)
	AdminListPendingKYC(ctx context.Context, page, pageSize int) (*entity.ListDriverProfilesResult, error)
	AdminReviewKYC(ctx context.Context, param *entity.ReviewKYCParam) (*entity.DriverProfile, error)
	CountPendingKYC(ctx context.Context) (int64, error)
}

type complianceImpl struct {
	repo ComplianceRepository
}

func NewCompliance(repo ComplianceRepository) ComplianceUseCase {
	return &complianceImpl{repo: repo}
}

func (c *complianceImpl) UpdateDriverKYC(ctx context.Context, param *entity.UpdateDriverKYCParam) (*entity.DriverProfile, error) {
	if param.UserID == uuid.Nil {
		return nil, entity.ErrInvalidUserID
	}
	if !entity.IsValidKycStatus(param.KycStatus) {
		return nil, entity.ErrInvalidKycStatus.WithDetail("kyc_status", param.KycStatus)
	}
	if err := c.mustBeDriver(ctx, param.UserID); err != nil {
		return nil, err
	}
	return c.repo.UpdateDriverKYC(ctx, param)
}

func (c *complianceImpl) AdminListPendingKYC(ctx context.Context, page, pageSize int) (*entity.ListDriverProfilesResult, error) {
	page, pageSize, _ = entity.NormalizePaging(page, pageSize)
	list, total, err := c.repo.ListPendingKYC(ctx, page, pageSize)
	if err != nil {
		return nil, err
	}
	return &entity.ListDriverProfilesResult{
		DriverProfiles: list,
		Pagination:     entity.BuildPagination(page, pageSize, total),
	}, nil
}

func (c *complianceImpl) AdminReviewKYC(ctx context.Context, param *entity.ReviewKYCParam) (*entity.DriverProfile, error) {
	if param.UserID == uuid.Nil {
		return nil, entity.ErrInvalidUserID
	}

	current, err := c.repo.GetDriverProfile(ctx, param.UserID)
	if err != nil {
		return nil, err
	}

	// Máy trạng thái: chỉ hồ sơ đang chờ mới duyệt được. Duyệt hai lần là lỗi
	// nghiệp vụ, không phải thao tác vô hại.
	if current.KycStatus != entity.KycPending {
		return nil, entity.ErrKycAlreadyReviewed.WithDetail("current_status", current.KycStatus)
	}

	status := entity.KycRejected
	if param.Approved {
		status = entity.KycApproved
	}

	return c.repo.UpdateDriverKYC(ctx, &entity.UpdateDriverKYCParam{
		UserID:     param.UserID,
		KycStatus:  status,
		Note:       param.Note,
		ReviewerID: param.ReviewerID,
	})
}

func (c *complianceImpl) CountPendingKYC(ctx context.Context) (int64, error) {
	return c.repo.CountPendingKYC(ctx)
}

func (c *complianceImpl) mustBeDriver(ctx context.Context, userID uuid.UUID) error {
	u, err := c.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if u.Role != entity.RoleDriver {
		return entity.ErrNotADriver.WithDetail("actual_role", u.Role)
	}
	return nil
}
