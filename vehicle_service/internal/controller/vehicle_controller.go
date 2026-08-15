package controller

import (
	"context"

	"vehicle_service/internal/biz"
	"vehicle_service/internal/mapper"
	"vehicle_service/internal/entity"
	vehiclev1 "github.com/logistic/api/logistic/vehicle_service/v1"
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
	param := &entity.RegisterVehicleParam{
		DriverID:          req.DriverId,
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

	return &vehiclev1.RegisterVehicleResponse{
		Id:      res.ID,
		Message: res.Message,
	}, nil
}

func (c *vehicleController) GetVehicle(ctx context.Context, req *vehiclev1.GetVehicleRequest) (*vehiclev1.GetVehicleResponse, error) {
	param := &entity.GetVehicleParam{
		ID: req.Id,
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
	param := &entity.ListVehiclesParam{
		DriverID: req.DriverId,
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
	param := &entity.UpdateVehicleStatusParam{
		ID:     req.Id,
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


