package biz

import (
	"context"
	"fmt"

	"vehicle_service/ent"
	"vehicle_service/ent/vehicle"
	"vehicle_service/internal/entity"

	"github.com/google/uuid"
)

type VehicleEngine interface {
	RegisterVehicle(ctx context.Context, param *entity.RegisterVehicleParam) (*entity.RegisterVehicleResult, error)
	GetVehicle(ctx context.Context, param *entity.GetVehicleParam) (*entity.Vehicle, error)
	ListVehicles(ctx context.Context, param *entity.ListVehiclesParam) ([]*entity.Vehicle, error)
	UpdateVehicleStatus(ctx context.Context, param *entity.UpdateVehicleStatusParam) (*entity.UpdateVehicleStatusResult, error)
}

type vehicleEngineImpl struct {
	repo VehicleRepo
}

func NewVehicleEngine(repo VehicleRepo) VehicleEngine {
	return &vehicleEngineImpl{repo: repo}
}

func (e *vehicleEngineImpl) RegisterVehicle(ctx context.Context, param *entity.RegisterVehicleParam) (*entity.RegisterVehicleResult, error) {
	driverID, err := uuid.Parse(param.DriverID)
	if err != nil {
		return nil, fmt.Errorf("invalid driver ID")
	}

	v := &ent.Vehicle{
		DriverID:          driverID,
		LicensePlate:      param.LicensePlate,
		Brand:             param.Brand,
		Model:             param.Model,
		VehicleType:       vehicle.VehicleType(param.VehicleType),
		CapacityWeightKg:  param.CapacityWeightKg,
		CapacityVolumeCbm: param.CapacityVolumeCbm,
	}

	created, err := e.repo.CreateVehicle(ctx, v)
	if err != nil {
		return nil, fmt.Errorf("failed to register vehicle: %w", err)
	}

	return &entity.RegisterVehicleResult{
		ID:      created.ID.String(),
		Message: "Vehicle registered successfully",
	}, nil
}

func (e *vehicleEngineImpl) GetVehicle(ctx context.Context, param *entity.GetVehicleParam) (*entity.Vehicle, error) {
	vid, err := uuid.Parse(param.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid vehicle ID")
	}

	v, err := e.repo.GetVehicleByID(ctx, vid)
	if err != nil {
		return nil, fmt.Errorf("vehicle not found")
	}

	return mapEntToEntity(v), nil
}

func (e *vehicleEngineImpl) ListVehicles(ctx context.Context, param *entity.ListVehiclesParam) ([]*entity.Vehicle, error) {
	driverID, err := uuid.Parse(param.DriverID)
	if err != nil {
		return nil, fmt.Errorf("invalid driver ID")
	}

	vehicles, err := e.repo.ListVehiclesByDriverID(ctx, driverID)
	if err != nil {
		return nil, fmt.Errorf("failed to list vehicles: %w", err)
	}

	var result []*entity.Vehicle
	for _, v := range vehicles {
		result = append(result, mapEntToEntity(v))
	}

	return result, nil
}

func (e *vehicleEngineImpl) UpdateVehicleStatus(ctx context.Context, param *entity.UpdateVehicleStatusParam) (*entity.UpdateVehicleStatusResult, error) {
	vid, err := uuid.Parse(param.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid vehicle ID")
	}

	if err := e.repo.UpdateVehicleStatus(ctx, vid, param.Status); err != nil {
		return nil, fmt.Errorf("failed to update status: %w", err)
	}

	return &entity.UpdateVehicleStatusResult{
		Message: "Vehicle status updated successfully",
	}, nil
}

func mapEntToEntity(v *ent.Vehicle) *entity.Vehicle {
	return &entity.Vehicle{
		ID:                v.ID,
		DriverID:          v.DriverID,
		LicensePlate:      v.LicensePlate,
		Brand:             v.Brand,
		Model:             v.Model,
		VehicleType:       string(v.VehicleType),
		CapacityWeightKg:  v.CapacityWeightKg,
		CapacityVolumeCbm: v.CapacityVolumeCbm,
		Status:            string(v.Status),
		CreatedAt:         v.CreatedAt,
		UpdatedAt:         v.UpdatedAt,
	}
}
