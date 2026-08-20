package biz

import (
	"context"

	"user_service/internal/entity"

	"github.com/google/uuid"
)

type UserRepo interface {
	CreateUser(ctx context.Context, u *entity.User) (*entity.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	GetUserByPhone(ctx context.Context, phone string) (*entity.User, error)
	ExistsByPhone(ctx context.Context, phone string) (bool, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	UpdateUser(ctx context.Context, param *entity.UpdateUserParam) (*entity.User, error)
	UpdateUserStatus(ctx context.Context, id uuid.UUID, status, reason string) (*entity.User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
	ListUsers(ctx context.Context, filter *entity.ListUsersFilter) ([]entity.User, int64, error)
	CountUsers(ctx context.Context, role, status string) (int64, error)

	CreateDriverProfile(ctx context.Context, userID uuid.UUID, dp *entity.DriverProfile) (*entity.DriverProfile, error)
	GetDriverProfile(ctx context.Context, userID uuid.UUID) (*entity.DriverProfile, error)
	UpdateDriverProfile(ctx context.Context, param *entity.UpdateDriverProfileParam) (*entity.DriverProfile, error)
	UpdateDriverKYC(ctx context.Context, param *entity.UpdateDriverKYCParam) (*entity.DriverProfile, error)
	ListPendingKYC(ctx context.Context, page, pageSize int) ([]entity.DriverProfile, int64, error)
	CountPendingKYC(ctx context.Context) (int64, error)

	CreateShipperProfile(ctx context.Context, userID uuid.UUID, sp *entity.ShipperProfile) (*entity.ShipperProfile, error)
	GetShipperProfile(ctx context.Context, userID uuid.UUID) (*entity.ShipperProfile, error)
	UpdateShipperProfile(ctx context.Context, param *entity.UpdateShipperProfileParam) (*entity.ShipperProfile, error)

	CreateAddress(ctx context.Context, param *entity.CreateAddressParam) (*entity.Address, error)
	GetAddress(ctx context.Context, id uuid.UUID) (*entity.Address, error)
	ListAddresses(ctx context.Context, param *entity.ListAddressesParam) ([]entity.Address, int64, error)
	UpdateAddress(ctx context.Context, param *entity.UpdateAddressParam) (*entity.Address, error)
	DeleteAddress(ctx context.Context, id uuid.UUID) error

	ClearDefaultAddress(ctx context.Context, userID uuid.UUID) error

	UpsertDevice(ctx context.Context, param *entity.RegisterDeviceParam) (*entity.UserDevice, error)
	GetDevice(ctx context.Context, id uuid.UUID) (*entity.UserDevice, error)
	ListDevices(ctx context.Context, userID uuid.UUID) ([]entity.UserDevice, error)
	DeleteDevice(ctx context.Context, id uuid.UUID) error
}