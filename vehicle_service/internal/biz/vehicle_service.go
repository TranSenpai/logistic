package biz

import "context"

type VehicleUsecase interface {
	CreateVehicle(ctx context.Context)
}
