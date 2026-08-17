package controller

import (
	"context"

	"vehicle_service/internal/biz"
	"vehicle_service/internal/mapper"
	"vehicle_service/internal/entity"
	vehiclev1 "github.com/logistic/api/logistic/vehicle_service/v1"
	"github.com/google/uuid"
)

type vehicleController struct {
	vehiclev1.UnimplementedVehicleServiceServer
	engine biz.VehicleEngine
	mapper mapper.AppMapper
}

func NewVehicleController(engine biz.VehicleEngine, appMapper mapper.AppMapper) vehiclev1.VehicleServiceServer {
	return &vehicleController{
		engine: engine,
		mapper: appMapper,
	}
}

func (c *vehicleController) RegisterVehicle(ctx context.Context, req *vehiclev1.RegisterVehicleRequest) (*vehiclev1.RegisterVehicleResponse, error) {
	parsedDriverID, _ := uuid.FromBytes(req.DriverId)
	param := &entity.RegisterVehicleParam{
		DriverID:          parsedDriverID.String(),
		LicensePlate:      req.LicensePlate,
		Brand:             req.Brand,
		Model:             req.Model,
		VehicleType:       req.VehicleType,
		CapacityWeightKg:  float64(req.CapacityWeightKg),
		CapacityVolumeCbm: float64(req.CapacityVolumeCbm),
	}

	res, err := c.engine.RegisterVehicle(ctx, param)
	if err != nil {
		return nil, err
	}

	// Convert string ID to bytes
	parsedID, _ := uuid.Parse(res.ID)

	return &vehiclev1.RegisterVehicleResponse{
		Id:      parsedID[:],
		Message: res.Message,
	}, nil
}

func (c *vehicleController) GetVehicle(ctx context.Context, req *vehiclev1.GetVehicleRequest) (*vehiclev1.GetVehicleResponse, error) {
	parsedID, _ := uuid.FromBytes(req.Id)
	param := &entity.GetVehicleParam{
		ID: parsedID.String(),
	}

	v, err := c.engine.GetVehicle(ctx, param)
	if err != nil {
		return nil, err
	}

	return &vehiclev1.GetVehicleResponse{
		Vehicle: c.mapper.EntityVehicleToPbVehicle(*v),
	}, nil
}

func (c *vehicleController) ListVehicles(ctx context.Context, req *vehiclev1.ListVehiclesRequest) (*vehiclev1.ListVehiclesResponse, error) {
	parsedDriverID, _ := uuid.FromBytes(req.DriverId)
	param := &entity.ListVehiclesParam{
		DriverID: parsedDriverID.String(),
	}

	vehicles, err := c.engine.ListVehicles(ctx, param)
	if err != nil {
		return nil, err
	}

	// Because ListVehicles returns []*entity.Vehicle but mapper expects []entity.Vehicle
	// Let's just map manually using the mapper for single entity
	var pbVehicles []*vehiclev1.Vehicle
	for _, v := range vehicles {
		pbVehicles = append(pbVehicles, c.mapper.EntityVehicleToPbVehicle(*v))
	}

	return &vehiclev1.ListVehiclesResponse{
		Vehicles: pbVehicles,
	}, nil
}

func (c *vehicleController) UpdateVehicleStatus(ctx context.Context, req *vehiclev1.UpdateVehicleStatusRequest) (*vehiclev1.UpdateVehicleStatusResponse, error) {
	parsedID, _ := uuid.FromBytes(req.Id)
	param := &entity.UpdateVehicleStatusParam{
		ID:     parsedID.String(),
		Status: req.Status,
	}

	res, err := c.engine.UpdateVehicleStatus(ctx, param)
	if err != nil {
		return nil, err
	}

	return &vehiclev1.UpdateVehicleStatusResponse{
		Message: res.Message,
	}, nil
}


