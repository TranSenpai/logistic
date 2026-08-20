// Package repo là tầng duy nhất được phép chạm vào ent (dao) và Redis.
//
// Chiến lược cache dùng ở đây là cache-aside + invalidate-on-write:
//
//	ĐỌC : thử Redis -> miss thì xuống Postgres -> ghi ngược lên Redis kèm TTL.
//	GHI : ghi Postgres trước, XOÁ key liên quan sau (không ghi đè cache).
//
// Vì sao xoá chứ không ghi đè? Ghi đè tạo ra khoảng thời gian hai request đồng
// thời có thể ghi lộn thứ tự, để lại bản cũ đè lên bản mới trong cache. Xoá thì
// tệ nhất chỉ là một lần cache miss — luôn đúng, chỉ tốn thêm một truy vấn.
//
// TTL vẫn được đặt kể cả khi đã invalidate chủ động: nếu một luồng ghi nào đó
// quên xoá key, dữ liệu bẩn cũng tự hết hạn sau vài phút thay vì nằm lại mãi.
package repo

import (
	"context"
	"log"
	"strings"
	"time"

	"user_service/ent"
	"user_service/ent/address"
	"user_service/ent/driverprofile"
	"user_service/ent/shipperprofile"
	"user_service/ent/user"
	"user_service/ent/userdevice"
	"user_service/internal/biz"
	cerr "user_service/internal/common/errors"
	"user_service/internal/entity"
	"user_service/internal/mapper"

	"github.com/google/uuid"
	"github.com/logistic/pkg/cache"
)

const (
	ttlUser     = 10 * time.Minute
	ttlProfile  = 10 * time.Minute
	ttlAddress  = 5 * time.Minute
	ttlDeviceLs = 5 * time.Minute
)

type userRepoImpl struct {
	client *ent.Client
	cache  *cache.Client
	mapper mapper.AppMapper
}

var _ biz.UserRepo = (*userRepoImpl)(nil)

// NewUserRepo nhận cache có thể là nil — khi Redis không dựng được, service vẫn
// chạy bình thường, chỉ là mọi truy vấn đều xuống thẳng Postgres.
func NewUserRepo(client *ent.Client, redis *cache.Client, appMapper mapper.AppMapper) biz.UserRepo {
	return &userRepoImpl{client: client, cache: redis, mapper: appMapper}
}

// ---------------------------------------------------------------------------
// KHOÁ CACHE
// ---------------------------------------------------------------------------

func (r *userRepoImpl) keyUser(id uuid.UUID) string    { return r.cache.Key("user", id.String()) }
func (r *userRepoImpl) keyPhone(phone string) string   { return r.cache.Key("phone", phone) }
func (r *userRepoImpl) keyDriver(id uuid.UUID) string  { return r.cache.Key("driver", id.String()) }
func (r *userRepoImpl) keyShipper(id uuid.UUID) string { return r.cache.Key("shipper", id.String()) }
func (r *userRepoImpl) keyAddrList(userID uuid.UUID) string {
	return r.cache.Key("addresses", userID.String())
}
func (r *userRepoImpl) keyDeviceList(userID uuid.UUID) string {
	return r.cache.Key("devices", userID.String())
}

// invalidateUser xoá mọi thứ liên quan tới một user sau khi ghi.
// Lỗi Redis ở đây chỉ log chứ không làm hỏng request: dữ liệu trong Postgres đã
// đúng rồi, cache bẩn nhiều nhất chỉ sống tới khi hết TTL.
func (r *userRepoImpl) invalidateUser(ctx context.Context, id uuid.UUID, phone string) {
	if r.cache == nil {
		return
	}
	keys := []string{r.keyUser(id), r.keyDriver(id), r.keyShipper(id)}
	if phone != "" {
		keys = append(keys, r.keyPhone(phone))
	}
	if err := r.cache.Delete(ctx, keys...); err != nil {
		log.Printf("[repo] invalidate user %s failed: %v", id, err)
	}
}

// ---------------------------------------------------------------------------
// USERS
// ---------------------------------------------------------------------------

func (r *userRepoImpl) CreateUser(ctx context.Context, u *entity.User) (*entity.User, error) {
	builder := r.client.User.Create().
		SetPhone(u.Phone).
		SetPasswordHash(u.PasswordHash).
		SetRole(user.Role(u.Role)).
		SetFullName(u.FullName)

	if u.Email != "" {
		builder = builder.SetEmail(u.Email)
	}
	if u.AvatarURL != "" {
		builder = builder.SetAvatarURL(u.AvatarURL)
	}

	dao, err := builder.Save(ctx)
	if err != nil {
		return nil, wrapError(err, cerr.ErrUserNotFound)
	}

	e := r.mapper.EntUserToEntityUser(dao)
	return &e, nil
}

func (r *userRepoImpl) GetUserByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	key := r.keyUser(id)

	var cached entity.User
	if r.cache != nil {
		if err := r.cache.GetJSON(ctx, key, &cached); err == nil {
			return &cached, nil
		}
	}

	dao, err := r.client.User.Query().Where(user.IDEQ(id)).Only(ctx)
	if err != nil {
		return nil, wrapError(err, cerr.ErrUserNotFound)
	}

	e := r.mapper.EntUserToEntityUser(dao)
	if r.cache != nil {
		_ = r.cache.SetJSON(ctx, key, e, ttlUser)
	}
	return &e, nil
}

func (r *userRepoImpl) GetUserByPhone(ctx context.Context, phone string) (*entity.User, error) {
	key := r.keyPhone(phone)

	var cached entity.User
	if r.cache != nil {
		if err := r.cache.GetJSON(ctx, key, &cached); err == nil {
			return &cached, nil
		}
	}

	dao, err := r.client.User.Query().Where(user.PhoneEQ(phone)).Only(ctx)
	if err != nil {
		return nil, wrapError(err, cerr.ErrUserNotFound)
	}

	e := r.mapper.EntUserToEntityUser(dao)
	if r.cache != nil {
		_ = r.cache.SetJSON(ctx, key, e, ttlUser)
	}
	return &e, nil
}

// ExistsByPhone dùng COUNT thay vì lấy cả bản ghi: rẻ hơn và không kéo
// password_hash ra khỏi DB một cách không cần thiết.
func (r *userRepoImpl) ExistsByPhone(ctx context.Context, phone string) (bool, error) {
	n, err := r.client.User.Query().Where(user.PhoneEQ(phone)).Count(ctx)
	if err != nil {
		return false, wrapError(err, cerr.ErrUserNotFound)
	}
	return n > 0, nil
}

func (r *userRepoImpl) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	if email == "" {
		return false, nil
	}
	n, err := r.client.User.Query().Where(user.EmailEQ(email)).Count(ctx)
	if err != nil {
		return false, wrapError(err, cerr.ErrUserNotFound)
	}
	return n > 0, nil
}

func (r *userRepoImpl) UpdateUser(ctx context.Context, param *entity.UpdateUserParam) (*entity.User, error) {
	builder := r.client.User.UpdateOneID(param.ID)

	// Chỉ ghi đè field mà client thực sự gửi lên. Nếu gán cả chuỗi rỗng thì một
	// request chỉ muốn đổi avatar sẽ vô tình xoá sạch họ tên của người dùng.
	if param.FullName != "" {
		builder = builder.SetFullName(param.FullName)
	}
	if param.Email != "" {
		builder = builder.SetEmail(param.Email)
	}
	if param.AvatarURL != "" {
		builder = builder.SetAvatarURL(param.AvatarURL)
	}

	dao, err := builder.Save(ctx)
	if err != nil {
		return nil, wrapError(err, cerr.ErrUserNotFound)
	}

	r.invalidateUser(ctx, dao.ID, dao.Phone)
	e := r.mapper.EntUserToEntityUser(dao)
	return &e, nil
}

func (r *userRepoImpl) UpdateUserStatus(ctx context.Context, id uuid.UUID, status, reason string) (*entity.User, error) {
	dao, err := r.client.User.UpdateOneID(id).
		SetStatus(user.Status(status)).
		SetStatusReason(reason).
		Save(ctx)
	if err != nil {
		return nil, wrapError(err, cerr.ErrUserNotFound)
	}

	r.invalidateUser(ctx, dao.ID, dao.Phone)
	e := r.mapper.EntUserToEntityUser(dao)
	return &e, nil
}

func (r *userRepoImpl) DeleteUser(ctx context.Context, id uuid.UUID) error {
	// Lấy phone trước khi xoá để còn biết key cache nào cần dọn.
	dao, err := r.client.User.Get(ctx, id)
	if err != nil {
		return wrapError(err, cerr.ErrUserNotFound)
	}

	// Xoá con trước rồi mới tới cha: Postgres sẽ chặn nếu còn bản ghi tham chiếu.
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return wrapError(err, nil)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Address.Delete().Where(address.UserIDEQ(id)).Exec(ctx); err != nil {
		return wrapError(err, cerr.ErrAddressNotFound)
	}
	if _, err := tx.UserDevice.Delete().Where(userdevice.UserIDEQ(id)).Exec(ctx); err != nil {
		return wrapError(err, cerr.ErrDeviceNotFound)
	}
	if _, err := tx.DriverProfile.Delete().Where(driverprofile.UserIDEQ(id)).Exec(ctx); err != nil {
		return wrapError(err, cerr.ErrDriverProfileNotFound)
	}
	if _, err := tx.ShipperProfile.Delete().Where(shipperprofile.UserIDEQ(id)).Exec(ctx); err != nil {
		return wrapError(err, cerr.ErrShipperProfileNotFound)
	}
	if err := tx.User.DeleteOneID(id).Exec(ctx); err != nil {
		return wrapError(err, cerr.ErrUserNotFound)
	}

	if err := tx.Commit(); err != nil {
		return wrapError(err, nil)
	}

	r.invalidateUser(ctx, id, dao.Phone)
	if r.cache != nil {
		_ = r.cache.Delete(ctx, r.keyAddrList(id), r.keyDeviceList(id))
	}
	return nil
}

// ListUsers KHÔNG cache: kết quả phụ thuộc tổ hợp filter + trang, số biến thể là
// vô hạn nên cache chỉ tổ làm đầy Redis mà tỉ lệ trúng gần bằng không.
func (r *userRepoImpl) ListUsers(ctx context.Context, filter *entity.ListUsersFilter) ([]entity.User, int64, error) {
	_, pageSize, offset := entity.NormalizePaging(filter.Page, filter.PageSize)

	q := r.client.User.Query()
	if filter.Role != "" {
		q = q.Where(user.RoleEQ(user.Role(filter.Role)))
	}
	if filter.Status != "" {
		q = q.Where(user.StatusEQ(user.Status(filter.Status)))
	}
	if kw := strings.TrimSpace(filter.Keyword); kw != "" {
		q = q.Where(user.Or(
			user.PhoneContainsFold(kw),
			user.EmailContainsFold(kw),
			user.FullNameContainsFold(kw),
		))
	}

	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, wrapError(err, cerr.ErrUserNotFound)
	}

	daos, err := q.
		Order(ent.Desc(user.FieldCreatedAt)).
		Offset(offset).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, 0, wrapError(err, cerr.ErrUserNotFound)
	}

	return r.mapper.EntUserListToEntityUserList(daos), int64(total), nil
}

func (r *userRepoImpl) CountUsers(ctx context.Context, role, status string) (int64, error) {
	q := r.client.User.Query()
	if role != "" {
		q = q.Where(user.RoleEQ(user.Role(role)))
	}
	if status != "" {
		q = q.Where(user.StatusEQ(user.Status(status)))
	}
	n, err := q.Count(ctx)
	if err != nil {
		return 0, wrapError(err, cerr.ErrUserNotFound)
	}
	return int64(n), nil
}

// ---------------------------------------------------------------------------
// DRIVER PROFILE
// ---------------------------------------------------------------------------

func (r *userRepoImpl) CreateDriverProfile(ctx context.Context, userID uuid.UUID, dp *entity.DriverProfile) (*entity.DriverProfile, error) {
	builder := r.client.DriverProfile.Create().SetUserID(userID)

	// Để NULL khi chưa có dữ liệu — xem ghi chú unique index trong ent/schema.
	if dp.LicenseNumber != "" {
		builder = builder.SetLicenseNumber(dp.LicenseNumber)
	}
	if dp.IDCard != "" {
		builder = builder.SetIDCard(dp.IDCard)
	}

	dao, err := builder.Save(ctx)
	if err != nil {
		return nil, wrapError(err, cerr.ErrDriverProfileNotFound)
	}

	e := r.mapper.EntDriverProfileToEntityDriverProfile(dao)
	return &e, nil
}

func (r *userRepoImpl) GetDriverProfile(ctx context.Context, userID uuid.UUID) (*entity.DriverProfile, error) {
	key := r.keyDriver(userID)

	var cached entity.DriverProfile
	if r.cache != nil {
		if err := r.cache.GetJSON(ctx, key, &cached); err == nil {
			return &cached, nil
		}
	}

	dao, err := r.client.DriverProfile.Query().Where(driverprofile.UserIDEQ(userID)).Only(ctx)
	if err != nil {
		return nil, wrapError(err, cerr.ErrDriverProfileNotFound)
	}

	e := r.mapper.EntDriverProfileToEntityDriverProfile(dao)
	if r.cache != nil {
		_ = r.cache.SetJSON(ctx, key, e, ttlProfile)
	}
	return &e, nil
}

func (r *userRepoImpl) UpdateDriverProfile(ctx context.Context, param *entity.UpdateDriverProfileParam) (*entity.DriverProfile, error) {
	dao, err := r.client.DriverProfile.Query().Where(driverprofile.UserIDEQ(param.UserID)).Only(ctx)
	if err != nil {
		return nil, wrapError(err, cerr.ErrDriverProfileNotFound)
	}

	builder := dao.Update()
	if param.LicenseNumber != "" {
		builder = builder.SetLicenseNumber(param.LicenseNumber)
	}
	if param.IDCard != "" {
		builder = builder.SetIDCard(param.IDCard)
	}

	updated, err := builder.Save(ctx)
	if err != nil {
		return nil, wrapError(err, cerr.ErrDriverProfileNotFound)
	}

	if r.cache != nil {
		_ = r.cache.Delete(ctx, r.keyDriver(param.UserID))
	}
	e := r.mapper.EntDriverProfileToEntityDriverProfile(updated)
	return &e, nil
}

func (r *userRepoImpl) UpdateDriverKYC(ctx context.Context, param *entity.UpdateDriverKYCParam) (*entity.DriverProfile, error) {
	dao, err := r.client.DriverProfile.Query().Where(driverprofile.UserIDEQ(param.UserID)).Only(ctx)
	if err != nil {
		return nil, wrapError(err, cerr.ErrDriverProfileNotFound)
	}

	builder := dao.Update().
		SetKycStatus(driverprofile.KycStatus(param.KycStatus)).
		SetKycNote(param.Note).
		SetKycReviewedAt(time.Now())

	if param.ReviewerID != uuid.Nil {
		builder = builder.SetKycReviewedBy(param.ReviewerID)
	}

	updated, err := builder.Save(ctx)
	if err != nil {
		return nil, wrapError(err, cerr.ErrDriverProfileNotFound)
	}

	if r.cache != nil {
		_ = r.cache.Delete(ctx, r.keyDriver(param.UserID))
	}
	e := r.mapper.EntDriverProfileToEntityDriverProfile(updated)
	return &e, nil
}

func (r *userRepoImpl) ListPendingKYC(ctx context.Context, page, pageSize int) ([]entity.DriverProfile, int64, error) {
	_, pageSize, offset := entity.NormalizePaging(page, pageSize)

	q := r.client.DriverProfile.Query().
		Where(driverprofile.KycStatusEQ(driverprofile.KycStatusPending))

	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, wrapError(err, cerr.ErrDriverProfileNotFound)
	}

	daos, err := q.
		Order(ent.Asc(driverprofile.FieldCreatedAt)). // hàng đợi duyệt: ai nộp trước xét trước
		Offset(offset).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, 0, wrapError(err, cerr.ErrDriverProfileNotFound)
	}

	return r.mapper.EntDriverProfileListToEntityList(daos), int64(total), nil
}

func (r *userRepoImpl) CountPendingKYC(ctx context.Context) (int64, error) {
	n, err := r.client.DriverProfile.Query().
		Where(driverprofile.KycStatusEQ(driverprofile.KycStatusPending)).
		Count(ctx)
	if err != nil {
		return 0, wrapError(err, cerr.ErrDriverProfileNotFound)
	}
	return int64(n), nil
}

// ---------------------------------------------------------------------------
// SHIPPER PROFILE
// ---------------------------------------------------------------------------

func (r *userRepoImpl) CreateShipperProfile(ctx context.Context, userID uuid.UUID, sp *entity.ShipperProfile) (*entity.ShipperProfile, error) {
	dao, err := r.client.ShipperProfile.Create().
		SetUserID(userID).
		SetCompanyName(sp.CompanyName).
		SetTaxCode(sp.TaxCode).
		SetBusinessAddress(sp.BusinessAddress).
		Save(ctx)
	if err != nil {
		return nil, wrapError(err, cerr.ErrShipperProfileNotFound)
	}

	e := r.mapper.EntShipperProfileToEntityShipperProfile(dao)
	return &e, nil
}

func (r *userRepoImpl) GetShipperProfile(ctx context.Context, userID uuid.UUID) (*entity.ShipperProfile, error) {
	key := r.keyShipper(userID)

	var cached entity.ShipperProfile
	if r.cache != nil {
		if err := r.cache.GetJSON(ctx, key, &cached); err == nil {
			return &cached, nil
		}
	}

	dao, err := r.client.ShipperProfile.Query().Where(shipperprofile.UserIDEQ(userID)).Only(ctx)
	if err != nil {
		return nil, wrapError(err, cerr.ErrShipperProfileNotFound)
	}

	e := r.mapper.EntShipperProfileToEntityShipperProfile(dao)
	if r.cache != nil {
		_ = r.cache.SetJSON(ctx, key, e, ttlProfile)
	}
	return &e, nil
}

func (r *userRepoImpl) UpdateShipperProfile(ctx context.Context, param *entity.UpdateShipperProfileParam) (*entity.ShipperProfile, error) {
	dao, err := r.client.ShipperProfile.Query().Where(shipperprofile.UserIDEQ(param.UserID)).Only(ctx)
	if err != nil {
		return nil, wrapError(err, cerr.ErrShipperProfileNotFound)
	}

	builder := dao.Update()
	if param.CompanyName != "" {
		builder = builder.SetCompanyName(param.CompanyName)
	}
	if param.TaxCode != "" {
		builder = builder.SetTaxCode(param.TaxCode)
	}
	if param.BusinessAddress != "" {
		builder = builder.SetBusinessAddress(param.BusinessAddress)
	}

	updated, err := builder.Save(ctx)
	if err != nil {
		return nil, wrapError(err, cerr.ErrShipperProfileNotFound)
	}

	if r.cache != nil {
		_ = r.cache.Delete(ctx, r.keyShipper(param.UserID))
	}
	e := r.mapper.EntShipperProfileToEntityShipperProfile(updated)
	return &e, nil
}

// ---------------------------------------------------------------------------
// ADDRESSES
// ---------------------------------------------------------------------------

func (r *userRepoImpl) CreateAddress(ctx context.Context, param *entity.CreateAddressParam) (*entity.Address, error) {
	dao, err := r.client.Address.Create().
		SetUserID(param.UserID).
		SetLabel(param.Label).
		SetContactName(param.ContactName).
		SetContactPhone(param.ContactPhone).
		SetLine1(param.Line1).
		SetWard(param.Ward).
		SetDistrict(param.District).
		SetCity(param.City).
		SetLatitude(param.Latitude).
		SetLongitude(param.Longitude).
		SetAddressType(address.AddressType(param.AddressType)).
		SetIsDefault(param.IsDefault).
		Save(ctx)
	if err != nil {
		return nil, wrapError(err, cerr.ErrAddressNotFound)
	}

	if r.cache != nil {
		_ = r.cache.Delete(ctx, r.keyAddrList(param.UserID))
	}
	e := r.mapper.EntAddressToEntityAddress(dao)
	return &e, nil
}

func (r *userRepoImpl) GetAddress(ctx context.Context, id uuid.UUID) (*entity.Address, error) {
	dao, err := r.client.Address.Get(ctx, id)
	if err != nil {
		return nil, wrapError(err, cerr.ErrAddressNotFound)
	}
	e := r.mapper.EntAddressToEntityAddress(dao)
	return &e, nil
}

func (r *userRepoImpl) ListAddresses(ctx context.Context, param *entity.ListAddressesParam) ([]entity.Address, int64, error) {
	page, pageSize, offset := entity.NormalizePaging(param.Page, param.PageSize)

	// Chỉ cache trang đầu không lọc — đó là thứ app gọi ở màn hình tạo đơn,
	// chiếm gần như toàn bộ lưu lượng của endpoint này.
	cacheable := r.cache != nil && param.AddressType == "" && page == 1
	if cacheable {
		var cached []entity.Address
		if err := r.cache.GetJSON(ctx, r.keyAddrList(param.UserID), &cached); err == nil {
			return cached, int64(len(cached)), nil
		}
	}

	q := r.client.Address.Query().Where(address.UserIDEQ(param.UserID))
	if param.AddressType != "" {
		q = q.Where(address.AddressTypeEQ(address.AddressType(param.AddressType)))
	}

	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, wrapError(err, cerr.ErrAddressNotFound)
	}

	daos, err := q.
		// Địa chỉ mặc định luôn đứng đầu để app chọn sẵn cho người dùng.
		Order(ent.Desc(address.FieldIsDefault), ent.Desc(address.FieldCreatedAt)).
		Offset(offset).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, 0, wrapError(err, cerr.ErrAddressNotFound)
	}

	list := r.mapper.EntAddressListToEntityAddressList(daos)
	if cacheable {
		_ = r.cache.SetJSON(ctx, r.keyAddrList(param.UserID), list, ttlAddress)
	}
	return list, int64(total), nil
}

func (r *userRepoImpl) UpdateAddress(ctx context.Context, param *entity.UpdateAddressParam) (*entity.Address, error) {
	builder := r.client.Address.UpdateOneID(param.ID).
		SetLatitude(param.Latitude).
		SetLongitude(param.Longitude).
		SetIsDefault(param.IsDefault)

	if param.Label != "" {
		builder = builder.SetLabel(param.Label)
	}
	if param.ContactName != "" {
		builder = builder.SetContactName(param.ContactName)
	}
	if param.ContactPhone != "" {
		builder = builder.SetContactPhone(param.ContactPhone)
	}
	if param.Line1 != "" {
		builder = builder.SetLine1(param.Line1)
	}
	if param.Ward != "" {
		builder = builder.SetWard(param.Ward)
	}
	if param.District != "" {
		builder = builder.SetDistrict(param.District)
	}
	if param.City != "" {
		builder = builder.SetCity(param.City)
	}
	if param.AddressType != "" {
		builder = builder.SetAddressType(address.AddressType(param.AddressType))
	}

	dao, err := builder.Save(ctx)
	if err != nil {
		return nil, wrapError(err, cerr.ErrAddressNotFound)
	}

	if r.cache != nil {
		_ = r.cache.Delete(ctx, r.keyAddrList(dao.UserID))
	}
	e := r.mapper.EntAddressToEntityAddress(dao)
	return &e, nil
}

func (r *userRepoImpl) DeleteAddress(ctx context.Context, id uuid.UUID) error {
	dao, err := r.client.Address.Get(ctx, id)
	if err != nil {
		return wrapError(err, cerr.ErrAddressNotFound)
	}
	if err := r.client.Address.DeleteOneID(id).Exec(ctx); err != nil {
		return wrapError(err, cerr.ErrAddressNotFound)
	}
	if r.cache != nil {
		_ = r.cache.Delete(ctx, r.keyAddrList(dao.UserID))
	}
	return nil
}

func (r *userRepoImpl) ClearDefaultAddress(ctx context.Context, userID uuid.UUID) error {
	_, err := r.client.Address.Update().
		Where(address.UserIDEQ(userID), address.IsDefaultEQ(true)).
		SetIsDefault(false).
		Save(ctx)
	if err != nil {
		return wrapError(err, cerr.ErrAddressNotFound)
	}
	if r.cache != nil {
		_ = r.cache.Delete(ctx, r.keyAddrList(userID))
	}
	return nil
}

// ---------------------------------------------------------------------------
// DEVICES
// ---------------------------------------------------------------------------

// UpsertDevice: cùng một device_token có thể được đăng ký lại sau khi người dùng
// đăng xuất rồi đăng nhập bằng tài khoản khác trên chính máy đó. Khi ấy phải
// CHUYỂN CHỦ bản ghi, không được tạo thêm dòng mới — nếu không thì push của tài
// khoản cũ vẫn bắn vào máy của người dùng mới.
func (r *userRepoImpl) UpsertDevice(ctx context.Context, param *entity.RegisterDeviceParam) (*entity.UserDevice, error) {
	existing, err := r.client.UserDevice.Query().
		Where(userdevice.DeviceTokenEQ(param.DeviceToken)).
		Only(ctx)

	if err == nil {
		updated, uErr := existing.Update().
			SetUserID(param.UserID).
			SetPlatform(userdevice.Platform(param.Platform)).
			SetDeviceName(param.DeviceName).
			SetIsActive(true).
			SetLastSeenAt(time.Now()).
			Save(ctx)
		if uErr != nil {
			return nil, wrapError(uErr, cerr.ErrDeviceNotFound)
		}

		if r.cache != nil {
			keys := []string{r.keyDeviceList(param.UserID)}
			if existing.UserID != param.UserID {
				keys = append(keys, r.keyDeviceList(existing.UserID))
			}
			_ = r.cache.Delete(ctx, keys...)
		}
		e := r.mapper.EntUserDeviceToEntityUserDevice(updated)
		return &e, nil
	}

	if !ent.IsNotFound(err) {
		return nil, wrapError(err, cerr.ErrDeviceNotFound)
	}

	dao, cErr := r.client.UserDevice.Create().
		SetUserID(param.UserID).
		SetDeviceToken(param.DeviceToken).
		SetPlatform(userdevice.Platform(param.Platform)).
		SetDeviceName(param.DeviceName).
		Save(ctx)
	if cErr != nil {
		return nil, wrapError(cErr, cerr.ErrDeviceNotFound)
	}

	if r.cache != nil {
		_ = r.cache.Delete(ctx, r.keyDeviceList(param.UserID))
	}
	e := r.mapper.EntUserDeviceToEntityUserDevice(dao)
	return &e, nil
}

func (r *userRepoImpl) GetDevice(ctx context.Context, id uuid.UUID) (*entity.UserDevice, error) {
	dao, err := r.client.UserDevice.Get(ctx, id)
	if err != nil {
		return nil, wrapError(err, cerr.ErrDeviceNotFound)
	}
	e := r.mapper.EntUserDeviceToEntityUserDevice(dao)
	return &e, nil
}

func (r *userRepoImpl) ListDevices(ctx context.Context, userID uuid.UUID) ([]entity.UserDevice, error) {
	key := r.keyDeviceList(userID)

	var cached []entity.UserDevice
	if r.cache != nil {
		if err := r.cache.GetJSON(ctx, key, &cached); err == nil {
			return cached, nil
		}
	}

	daos, err := r.client.UserDevice.Query().
		Where(userdevice.UserIDEQ(userID), userdevice.IsActiveEQ(true)).
		Order(ent.Desc(userdevice.FieldLastSeenAt)).
		All(ctx)
	if err != nil {
		return nil, wrapError(err, cerr.ErrDeviceNotFound)
	}

	list := r.mapper.EntUserDeviceListToEntityList(daos)
	if r.cache != nil {
		_ = r.cache.SetJSON(ctx, key, list, ttlDeviceLs)
	}
	return list, nil
}

func (r *userRepoImpl) DeleteDevice(ctx context.Context, id uuid.UUID) error {
	dao, err := r.client.UserDevice.Get(ctx, id)
	if err != nil {
		return wrapError(err, cerr.ErrDeviceNotFound)
	}
	if err := r.client.UserDevice.DeleteOneID(id).Exec(ctx); err != nil {
		return wrapError(err, cerr.ErrDeviceNotFound)
	}
	if r.cache != nil {
		_ = r.cache.Delete(ctx, r.keyDeviceList(dao.UserID))
	}
	return nil
}
