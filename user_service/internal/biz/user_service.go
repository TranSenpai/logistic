package biz

import (
	"context"

	cerr "user_service/internal/common/errors"
	"user_service/internal/entity"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserEngine interface {
	RegisterUser(ctx context.Context, param *entity.RegisterUserParam) (*entity.RegisterUserResult, error)
	GetUser(ctx context.Context, id uuid.UUID) (*entity.GetUserResult, error)
	UpdateUser(ctx context.Context, param *entity.UpdateUserParam) (*entity.User, error)

	GetDriverProfile(ctx context.Context, userID uuid.UUID) (*entity.DriverProfile, error)
	UpdateDriverProfile(ctx context.Context, param *entity.UpdateDriverProfileParam) (*entity.DriverProfile, error)
	GetShipperProfile(ctx context.Context, userID uuid.UUID) (*entity.ShipperProfile, error)
	UpdateShipperProfile(ctx context.Context, param *entity.UpdateShipperProfileParam) (*entity.ShipperProfile, error)
	UpdateDriverKYC(ctx context.Context, param *entity.UpdateDriverKYCParam) (*entity.DriverProfile, error)

	CreateAddress(ctx context.Context, param *entity.CreateAddressParam) (*entity.Address, error)
	ListAddresses(ctx context.Context, param *entity.ListAddressesParam) (*entity.ListAddressesResult, error)
	UpdateAddress(ctx context.Context, param *entity.UpdateAddressParam) (*entity.Address, error)
	DeleteAddress(ctx context.Context, id, userID uuid.UUID) error

	RegisterDevice(ctx context.Context, param *entity.RegisterDeviceParam) (*entity.UserDevice, error)
	ListDevices(ctx context.Context, userID uuid.UUID) ([]entity.UserDevice, error)
	DeleteDevice(ctx context.Context, id, userID uuid.UUID) error

	AdminListUsers(ctx context.Context, filter *entity.ListUsersFilter) (*entity.ListUsersResult, error)
	AdminUpdateUserStatus(ctx context.Context, param *entity.UpdateUserStatusParam) (*entity.User, error)
	AdminListPendingKYC(ctx context.Context, page, pageSize int) (*entity.ListDriverProfilesResult, error)
	AdminReviewKYC(ctx context.Context, param *entity.ReviewKYCParam) (*entity.DriverProfile, error)
	AdminGetUserStats(ctx context.Context) (*entity.UserStats, error)
	AdminDeleteUser(ctx context.Context, id uuid.UUID) error
}

type userEngineImpl struct {
	repo UserRepo
}

func NewUserEngine(repo UserRepo) UserEngine {
	return &userEngineImpl{repo: repo}
}

func (e *userEngineImpl) RegisterUser(ctx context.Context, param *entity.RegisterUserParam) (*entity.RegisterUserResult, error) {
	if !entity.IsValidRole(param.Role) {
		return nil, cerr.ErrInvalidRole.WithDetail("role", param.Role)
	}

	if !param.ProvisionedFromAuth() {
		if param.Phone == "" {
			return nil, cerr.ErrPhoneRequired
		}
		if len(param.Password) < 6 {
			return nil, cerr.ErrPasswordTooShort
		}
	}

	if param.Phone != "" {
		exists, err := e.repo.ExistsByPhone(ctx, param.Phone)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, cerr.ErrPhoneAlreadyUsed.WithDetail("phone", param.Phone)
		}
	}

	if param.Email != "" {
		emailTaken, eErr := e.repo.ExistsByEmail(ctx, param.Email)
		if eErr != nil {
			return nil, eErr
		}
		if emailTaken {
			return nil, cerr.ErrEmailAlreadyUsed.WithDetail("email", param.Email)
		}
	}

	// Không giữ bản sao credential thứ hai.
	passwordHash := ""
	if !param.ProvisionedFromAuth() {
		hashed, hErr := bcrypt.GenerateFromPassword([]byte(param.Password), bcrypt.DefaultCost)
		if hErr != nil {
			return nil, cerr.ErrDatabase.WithCause(hErr)
		}
		passwordHash = string(hashed)
	}

	created, err := e.repo.CreateUser(ctx, &entity.User{
		ID:           param.ID,
		Phone:        param.Phone,
		Email:        param.Email,
		FullName:     param.FullName,
		PasswordHash: passwordHash,
		Role:         param.Role,
	})
	if err != nil {
		return nil, err
	}

	switch param.Role {
	case entity.RoleDriver:
		if _, err := e.repo.CreateDriverProfile(ctx, created.ID, &entity.DriverProfile{}); err != nil {
			return nil, err
		}
	case entity.RoleShipper:
		if _, err := e.repo.CreateShipperProfile(ctx, created.ID, &entity.ShipperProfile{}); err != nil {
			return nil, err
		}
	}

	return &entity.RegisterUserResult{
		ID:      created.ID,
		User:    created,
		Message: "Đăng ký người dùng thành công",
	}, nil
}

func (e *userEngineImpl) GetUser(ctx context.Context, id uuid.UUID) (*entity.GetUserResult, error) {
	if id == uuid.Nil {
		return nil, cerr.ErrInvalidUserID
	}

	u, err := e.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	result := &entity.GetUserResult{User: u}

	switch u.Role {
	case entity.RoleDriver:
		if dp, dErr := e.repo.GetDriverProfile(ctx, id); dErr == nil {
			result.DriverProfile = dp
		}
	case entity.RoleShipper:
		if sp, sErr := e.repo.GetShipperProfile(ctx, id); sErr == nil {
			result.ShipperProfile = sp
		}
	}

	return result, nil
}

func (e *userEngineImpl) UpdateUser(ctx context.Context, param *entity.UpdateUserParam) (*entity.User, error) {
	if param.ID == uuid.Nil {
		return nil, cerr.ErrInvalidUserID
	}
	if param.Email != "" {
		taken, err := e.repo.ExistsByEmail(ctx, param.Email)
		if err != nil {
			return nil, err
		}
		if taken {
			current, cErr := e.repo.GetUserByID(ctx, param.ID)
			if cErr != nil {
				return nil, cErr
			}
			if current.Email != param.Email {
				return nil, cerr.ErrEmailAlreadyUsed.WithDetail("email", param.Email)
			}
		}
	}
	return e.repo.UpdateUser(ctx, param)
}

func (e *userEngineImpl) GetDriverProfile(ctx context.Context, userID uuid.UUID) (*entity.DriverProfile, error) {
	if userID == uuid.Nil {
		return nil, cerr.ErrInvalidUserID
	}
	if err := e.mustBeRole(ctx, userID, entity.RoleDriver); err != nil {
		return nil, err
	}
	return e.repo.GetDriverProfile(ctx, userID)
}

func (e *userEngineImpl) UpdateDriverProfile(ctx context.Context, param *entity.UpdateDriverProfileParam) (*entity.DriverProfile, error) {
	if param.UserID == uuid.Nil {
		return nil, cerr.ErrInvalidUserID
	}
	if err := e.mustBeRole(ctx, param.UserID, entity.RoleDriver); err != nil {
		return nil, err
	}
	return e.repo.UpdateDriverProfile(ctx, param)
}

func (e *userEngineImpl) GetShipperProfile(ctx context.Context, userID uuid.UUID) (*entity.ShipperProfile, error) {
	if userID == uuid.Nil {
		return nil, cerr.ErrInvalidUserID
	}
	if err := e.mustBeRole(ctx, userID, entity.RoleShipper); err != nil {
		return nil, err
	}
	return e.repo.GetShipperProfile(ctx, userID)
}

func (e *userEngineImpl) UpdateShipperProfile(ctx context.Context, param *entity.UpdateShipperProfileParam) (*entity.ShipperProfile, error) {
	if param.UserID == uuid.Nil {
		return nil, cerr.ErrInvalidUserID
	}
	if err := e.mustBeRole(ctx, param.UserID, entity.RoleShipper); err != nil {
		return nil, err
	}
	return e.repo.UpdateShipperProfile(ctx, param)
}

func (e *userEngineImpl) UpdateDriverKYC(ctx context.Context, param *entity.UpdateDriverKYCParam) (*entity.DriverProfile, error) {
	if param.UserID == uuid.Nil {
		return nil, cerr.ErrInvalidUserID
	}
	if !entity.IsValidKycStatus(param.KycStatus) {
		return nil, cerr.ErrInvalidKycStatus.WithDetail("kyc_status", param.KycStatus)
	}
	if err := e.mustBeRole(ctx, param.UserID, entity.RoleDriver); err != nil {
		return nil, err
	}
	return e.repo.UpdateDriverKYC(ctx, param)
}

func (e *userEngineImpl) mustBeRole(ctx context.Context, userID uuid.UUID, role string) error {
	u, err := e.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if u.Role != role {
		if role == entity.RoleDriver {
			return cerr.ErrNotADriver.WithDetail("actual_role", u.Role)
		}
		return cerr.ErrNotAShipper.WithDetail("actual_role", u.Role)
	}
	return nil
}

func (e *userEngineImpl) CreateAddress(ctx context.Context, param *entity.CreateAddressParam) (*entity.Address, error) {
	if param.UserID == uuid.Nil {
		return nil, cerr.ErrInvalidUserID
	}
	if param.Line1 == "" {
		return nil, cerr.ErrAddressRequired
	}
	if param.AddressType == "" {
		param.AddressType = entity.AddressTypeBoth
	}
	if !entity.IsValidAddressType(param.AddressType) {
		return nil, cerr.ErrInvalidAddrType.WithDetail("address_type", param.AddressType)
	}

	if _, err := e.repo.GetUserByID(ctx, param.UserID); err != nil {
		return nil, err
	}

	if param.IsDefault {
		if err := e.repo.ClearDefaultAddress(ctx, param.UserID); err != nil {
			return nil, err
		}
	}

	return e.repo.CreateAddress(ctx, param)
}

func (e *userEngineImpl) ListAddresses(ctx context.Context, param *entity.ListAddressesParam) (*entity.ListAddressesResult, error) {
	if param.UserID == uuid.Nil {
		return nil, cerr.ErrInvalidUserID
	}
	if param.AddressType != "" && !entity.IsValidAddressType(param.AddressType) {
		return nil, cerr.ErrInvalidAddrType.WithDetail("address_type", param.AddressType)
	}

	page, pageSize, _ := entity.NormalizePaging(param.Page, param.PageSize)
	list, total, err := e.repo.ListAddresses(ctx, param)
	if err != nil {
		return nil, err
	}

	return &entity.ListAddressesResult{
		Addresses:  list,
		Pagination: entity.BuildPagination(page, pageSize, total),
	}, nil
}

func (e *userEngineImpl) UpdateAddress(ctx context.Context, param *entity.UpdateAddressParam) (*entity.Address, error) {
	if param.ID == uuid.Nil {
		return nil, cerr.ErrInvalidAddressID
	}
	if param.AddressType != "" && !entity.IsValidAddressType(param.AddressType) {
		return nil, cerr.ErrInvalidAddrType.WithDetail("address_type", param.AddressType)
	}

	current, err := e.repo.GetAddress(ctx, param.ID)
	if err != nil {
		return nil, err
	}

	if param.UserID != uuid.Nil && current.UserID != param.UserID {
		return nil, cerr.ErrAddressNotOwned
	}

	if param.IsDefault && !current.IsDefault {
		if err := e.repo.ClearDefaultAddress(ctx, current.UserID); err != nil {
			return nil, err
		}
	}

	return e.repo.UpdateAddress(ctx, param)
}

func (e *userEngineImpl) DeleteAddress(ctx context.Context, id, userID uuid.UUID) error {
	if id == uuid.Nil {
		return cerr.ErrInvalidAddressID
	}
	current, err := e.repo.GetAddress(ctx, id)
	if err != nil {
		return err
	}
	if userID != uuid.Nil && current.UserID != userID {
		return cerr.ErrAddressNotOwned
	}
	return e.repo.DeleteAddress(ctx, id)
}

func (e *userEngineImpl) RegisterDevice(ctx context.Context, param *entity.RegisterDeviceParam) (*entity.UserDevice, error) {
	if param.UserID == uuid.Nil {
		return nil, cerr.ErrInvalidUserID
	}
	if param.DeviceToken == "" {
		return nil, cerr.ErrDeviceTokenEmpty
	}
	if param.Platform == "" {
		param.Platform = entity.PlatformAndroid
	}
	if !entity.IsValidPlatform(param.Platform) {
		return nil, cerr.ErrInvalidPlatform.WithDetail("platform", param.Platform)
	}

	if _, err := e.repo.GetUserByID(ctx, param.UserID); err != nil {
		return nil, err
	}
	return e.repo.UpsertDevice(ctx, param)
}

func (e *userEngineImpl) ListDevices(ctx context.Context, userID uuid.UUID) ([]entity.UserDevice, error) {
	if userID == uuid.Nil {
		return nil, cerr.ErrInvalidUserID
	}
	return e.repo.ListDevices(ctx, userID)
}

func (e *userEngineImpl) DeleteDevice(ctx context.Context, id, userID uuid.UUID) error {
	if id == uuid.Nil {
		return cerr.ErrInvalidDeviceID
	}
	current, err := e.repo.GetDevice(ctx, id)
	if err != nil {
		return err
	}
	if userID != uuid.Nil && current.UserID != userID {
		return cerr.ErrDeviceNotOwned
	}
	return e.repo.DeleteDevice(ctx, id)
}

func (e *userEngineImpl) AdminListUsers(ctx context.Context, filter *entity.ListUsersFilter) (*entity.ListUsersResult, error) {
	if filter.Role != "" && !entity.IsValidRole(filter.Role) {
		return nil, cerr.ErrInvalidRole.WithDetail("role", filter.Role)
	}
	if filter.Status != "" && !entity.IsValidStatus(filter.Status) {
		return nil, cerr.ErrInvalidStatus.WithDetail("status", filter.Status)
	}

	page, pageSize, _ := entity.NormalizePaging(filter.Page, filter.PageSize)
	users, total, err := e.repo.ListUsers(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &entity.ListUsersResult{
		Users:      users,
		Pagination: entity.BuildPagination(page, pageSize, total),
	}, nil
}

func (e *userEngineImpl) AdminUpdateUserStatus(ctx context.Context, param *entity.UpdateUserStatusParam) (*entity.User, error) {
	if param.ID == uuid.Nil {
		return nil, cerr.ErrInvalidUserID
	}
	if !entity.IsValidStatus(param.Status) {
		return nil, cerr.ErrInvalidStatus.WithDetail("status", param.Status)
	}
	return e.repo.UpdateUserStatus(ctx, param.ID, param.Status, param.Reason)
}

func (e *userEngineImpl) AdminListPendingKYC(ctx context.Context, page, pageSize int) (*entity.ListDriverProfilesResult, error) {
	page, pageSize, _ = entity.NormalizePaging(page, pageSize)
	list, total, err := e.repo.ListPendingKYC(ctx, page, pageSize)
	if err != nil {
		return nil, err
	}
	return &entity.ListDriverProfilesResult{
		DriverProfiles: list,
		Pagination:     entity.BuildPagination(page, pageSize, total),
	}, nil
}

func (e *userEngineImpl) AdminReviewKYC(ctx context.Context, param *entity.ReviewKYCParam) (*entity.DriverProfile, error) {
	if param.UserID == uuid.Nil {
		return nil, cerr.ErrInvalidUserID
	}

	current, err := e.repo.GetDriverProfile(ctx, param.UserID)
	if err != nil {
		return nil, err
	}

	if current.KycStatus != entity.KycPending {
		return nil, cerr.ErrKycAlreadyReviewed.WithDetail("current_status", current.KycStatus)
	}

	status := entity.KycRejected
	if param.Approved {
		status = entity.KycApproved
	}

	return e.repo.UpdateDriverKYC(ctx, &entity.UpdateDriverKYCParam{
		UserID:     param.UserID,
		KycStatus:  status,
		Note:       param.Note,
		ReviewerID: param.ReviewerID,
	})
}

func (e *userEngineImpl) AdminGetUserStats(ctx context.Context) (*entity.UserStats, error) {
	total, err := e.repo.CountUsers(ctx, "", "")
	if err != nil {
		return nil, err
	}
	drivers, err := e.repo.CountUsers(ctx, entity.RoleDriver, "")
	if err != nil {
		return nil, err
	}
	shippers, err := e.repo.CountUsers(ctx, entity.RoleShipper, "")
	if err != nil {
		return nil, err
	}
	active, err := e.repo.CountUsers(ctx, "", entity.StatusActive)
	if err != nil {
		return nil, err
	}
	banned, err := e.repo.CountUsers(ctx, "", entity.StatusBanned)
	if err != nil {
		return nil, err
	}
	pendingKyc, err := e.repo.CountPendingKYC(ctx)
	if err != nil {
		return nil, err
	}

	return &entity.UserStats{
		TotalUsers:    total,
		TotalDrivers:  drivers,
		TotalShippers: shippers,
		ActiveUsers:   active,
		BannedUsers:   banned,
		PendingKyc:    pendingKyc,
	}, nil
}

func (e *userEngineImpl) AdminDeleteUser(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return cerr.ErrInvalidUserID
	}
	return e.repo.DeleteUser(ctx, id)
}