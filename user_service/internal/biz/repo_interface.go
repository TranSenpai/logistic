package biz

import (
	"context"

	"user_service/ent"
	"github.com/google/uuid"
)

type UserRepo interface {
	CreateUser(ctx context.Context, u *ent.User) (*ent.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*ent.User, error)
	GetUserByPhone(ctx context.Context, phone string) (*ent.User, error)
	
	CreateDriverProfile(ctx context.Context, dp *ent.DriverProfile) error
	CreateShipperProfile(ctx context.Context, sp *ent.ShipperProfile) error

	GetDriverProfile(ctx context.Context, userID uuid.UUID) (*ent.DriverProfile, error)
	GetShipperProfile(ctx context.Context, userID uuid.UUID) (*ent.ShipperProfile, error)

	UpdateDriverKYC(ctx context.Context, userID uuid.UUID, status string) error
}
