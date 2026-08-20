// Package mapper khai báo HỢP ĐỒNG chuyển đổi giữa 3 tầng dữ liệu; phần thân
// hàm do goverter sinh ra trong package generated.
//
//	ent.*    (dao — do ent generate)      <->  entity.*  (viết tay)
//	entity.* (viết tay)                   <->  pb.*      (dto — do protobuf generate)
//
// Cách làm giống hệt matching_service: interface + comment directive, chạy
//
//	go generate ./internal/mapper
//
// là ra file generated/generated.go. Không ai được sửa tay file đó.
//
// Vì sao không gán field thủ công? Với 5 entity x 3 tầng, viết tay là khoảng
// 600 dòng gán lặp — và mỗi lần thêm cột lại phải nhớ sửa đủ 3 chỗ. Goverter
// bắt lỗi ngay lúc BIÊN DỊCH nếu có field không map được, nên quên là không
// build nổi chứ không phải chờ tới lúc chạy mới lòi ra field rỗng.
package mapper

import (
	"time"

	"github.com/google/uuid"
	pb "github.com/logistic/api/logistic/user_service/v1"
	"google.golang.org/protobuf/types/known/timestamppb"

	"user_service/ent"
	"user_service/ent/address"
	"user_service/ent/driverprofile"
	"user_service/ent/user"
	"user_service/ent/userdevice"
	"user_service/internal/entity"
)

// goverter:converter
// goverter:matchIgnoreCase
// goverter:useZeroValueOnPointerInconsistency
// goverter:ignoreUnexported
// goverter:extend IdentityTime
// goverter:extend UUIDToString
// goverter:extend StringToUUID
// goverter:extend TimeToTimestamp
// goverter:extend TimestampToTime
// goverter:extend StringPtrToString
// goverter:extend IntToInt32
// goverter:extend Int32ToInt
// goverter:extend EntUserRoleToString
// goverter:extend EntUserStatusToString
// goverter:extend EntKycStatusToString
// goverter:extend EntAddressTypeToString
// goverter:extend EntPlatformToString
//
//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.4 gen ./
type AppMapper interface {
	// ==================== DAO -> ENTITY (tầng repo) ====================

	EntUserToEntityUser(source *ent.User) entity.User
	EntUserListToEntityUserList(source []*ent.User) []entity.User

	EntDriverProfileToEntityDriverProfile(source *ent.DriverProfile) entity.DriverProfile
	EntDriverProfileListToEntityList(source []*ent.DriverProfile) []entity.DriverProfile

	EntShipperProfileToEntityShipperProfile(source *ent.ShipperProfile) entity.ShipperProfile

	EntAddressToEntityAddress(source *ent.Address) entity.Address
	EntAddressListToEntityAddressList(source []*ent.Address) []entity.Address

	EntUserDeviceToEntityUserDevice(source *ent.UserDevice) entity.UserDevice
	EntUserDeviceListToEntityList(source []*ent.UserDevice) []entity.UserDevice

	// ==================== ENTITY -> DTO (tầng controller) ====================

	EntityUserToPbUser(source entity.User) *pb.User
	EntityUserListToPbUserList(source []entity.User) []*pb.User

	EntityDriverProfileToPb(source entity.DriverProfile) *pb.DriverProfile
	EntityDriverProfileListToPbList(source []entity.DriverProfile) []*pb.DriverProfile

	EntityShipperProfileToPb(source entity.ShipperProfile) *pb.ShipperProfile

	EntityAddressToPb(source entity.Address) *pb.Address
	EntityAddressListToPbList(source []entity.Address) []*pb.Address

	EntityUserDeviceToPb(source entity.UserDevice) *pb.UserDevice
	EntityUserDeviceListToPbList(source []entity.UserDevice) []*pb.UserDevice

	EntityPaginationToPb(source entity.Pagination) *pb.Pagination

	// ==================== DTO -> ENTITY (tầng controller) ====================
	// Các hàm này trả (T, error) vì StringToUUID có thể thất bại khi client gửi
	// chuỗi không phải UUID. Lỗi được controller bọc lại thành ErrInvalidUserID.

	PbRegisterUserToParam(req *pb.RegisterUserRequest) entity.RegisterUserParam

	PbUpdateUserToParam(req *pb.UpdateUserRequest) (entity.UpdateUserParam, error)

	PbUpdateDriverProfileToParam(req *pb.UpdateDriverProfileRequest) (entity.UpdateDriverProfileParam, error)

	PbUpdateShipperProfileToParam(req *pb.UpdateShipperProfileRequest) (entity.UpdateShipperProfileParam, error)

	// goverter:ignore ReviewerID
	PbUpdateKycToParam(req *pb.UpdateDriverKYCRequest) (entity.UpdateDriverKYCParam, error)

	PbCreateAddressToParam(req *pb.CreateAddressRequest) (entity.CreateAddressParam, error)

	PbUpdateAddressToParam(req *pb.UpdateAddressRequest) (entity.UpdateAddressParam, error)

	PbListAddressesToParam(req *pb.ListAddressesRequest) (entity.ListAddressesParam, error)

	PbRegisterDeviceToParam(req *pb.RegisterDeviceRequest) (entity.RegisterDeviceParam, error)

	PbAdminListUsersToFilter(req *pb.AdminListUsersRequest) entity.ListUsersFilter

	PbAdminUpdateStatusToParam(req *pb.AdminUpdateUserStatusRequest) (entity.UpdateUserStatusParam, error)

	PbAdminReviewKycToParam(req *pb.AdminReviewKYCRequest) (entity.ReviewKYCParam, error)
}

// ===========================================================================
// HELPERS — được đăng ký qua `goverter:extend` ở trên nên goverter tự dùng
// chúng cho MỌI cặp kiểu tương ứng, không cần khai báo lại ở từng field.
// ===========================================================================

func IdentityTime(t time.Time) time.Time { return t }

func UUIDToString(u uuid.UUID) string {
	if u == uuid.Nil {
		return ""
	}
	return u.String()
}

// StringToUUID: chuỗi rỗng -> uuid.Nil (field không bắt buộc), chuỗi rác -> lỗi.
// Tầng biz sẽ tự từ chối uuid.Nil ở những chỗ bắt buộc phải có id.
func StringToUUID(s string) (uuid.UUID, error) {
	if s == "" {
		return uuid.Nil, nil
	}
	return uuid.Parse(s)
}

func TimeToTimestamp(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}

func TimestampToTime(ts *timestamppb.Timestamp) time.Time {
	if ts != nil && ts.IsValid() {
		return ts.AsTime()
	}
	return time.Time{}
}

// StringPtrToString: cột nullable của Postgres về Go là *string. Tầng nghiệp vụ
// không quan tâm phân biệt NULL với "" nên quy hết về chuỗi rỗng.
func StringPtrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func IntToInt32(i int) int32 { return int32(i) }
func Int32ToInt(i int32) int { return int(i) }

// Các enum của ent là kiểu string riêng; đưa về string thuần cho entity.
func EntUserRoleToString(r user.Role) string                { return string(r) }
func EntUserStatusToString(s user.Status) string            { return string(s) }
func EntKycStatusToString(s driverprofile.KycStatus) string { return string(s) }
func EntAddressTypeToString(t address.AddressType) string   { return string(t) }
func EntPlatformToString(p userdevice.Platform) string      { return string(p) }
