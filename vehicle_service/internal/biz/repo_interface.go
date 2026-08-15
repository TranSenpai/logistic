package biz

import (
	"context"

	"vehicle_service/ent"
	"github.com/google/uuid"
)

type VehicleRepo interface {
	CreateVehicle(ctx context.Context, v *ent.Vehicle) (*ent.Vehicle, error)
	GetVehicleByID(ctx context.Context, id uuid.UUID) (*ent.Vehicle, error)
	ListVehiclesByDriverID(ctx context.Context, driverID uuid.UUID) ([]*ent.Vehicle, error)
	UpdateVehicleStatus(ctx context.Context, id uuid.UUID, status string) error
}
