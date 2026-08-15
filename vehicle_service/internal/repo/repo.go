package repo

import (
	"context"

	"vehicle_service/ent"
	"vehicle_service/ent/vehicle"
	"github.com/google/uuid"
)

type vehicleRepoImpl struct {
	client *ent.Client
}

func NewVehicleRepo(client *ent.Client) *vehicleRepoImpl {
	return &vehicleRepoImpl{client: client}
}

func (r *vehicleRepoImpl) CreateVehicle(ctx context.Context, v *ent.Vehicle) (*ent.Vehicle, error) {
	return r.client.Vehicle.Create().
		SetDriverID(v.DriverID).
		SetLicensePlate(v.LicensePlate).
		SetNillableBrand(func() *string {
			if v.Brand != "" {
				return &v.Brand
			}
			return nil
		}()).
		SetNillableModel(func() *string {
			if v.Model != "" {
				return &v.Model
			}
			return nil
		}()).
		SetVehicleType(vehicle.VehicleType(v.VehicleType)).
		SetCapacityWeightKg(v.CapacityWeightKg).
		SetCapacityVolumeCbm(v.CapacityVolumeCbm).
		Save(ctx)
}

func (r *vehicleRepoImpl) GetVehicleByID(ctx context.Context, id uuid.UUID) (*ent.Vehicle, error) {
	return r.client.Vehicle.Query().Where(vehicle.IDEQ(id)).First(ctx)
}

func (r *vehicleRepoImpl) ListVehiclesByDriverID(ctx context.Context, driverID uuid.UUID) ([]*ent.Vehicle, error) {
	return r.client.Vehicle.Query().Where(vehicle.DriverIDEQ(driverID)).All(ctx)
}

func (r *vehicleRepoImpl) UpdateVehicleStatus(ctx context.Context, id uuid.UUID, status string) error {
	return r.client.Vehicle.UpdateOneID(id).
		SetStatus(vehicle.Status(status)).
		Exec(ctx)
}
