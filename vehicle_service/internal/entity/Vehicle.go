package entity

import (
	"time"

	"github.com/google/uuid"
)

type Vehicle struct {
	ID                uuid.UUID
	DriverID          uuid.UUID
	LicensePlate      string
	Brand             string
	Model             string
	VehicleType       string
	CapacityWeightKg  float64
	CapacityVolumeCbm float64
	Status            string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type RegisterVehicleParam struct {
	DriverID          string
	LicensePlate      string
	Brand             string
	Model             string
	VehicleType       string
	CapacityWeightKg  float64
	CapacityVolumeCbm float64
}

type RegisterVehicleResult struct {
	ID      string
	Message string
}

type GetVehicleParam struct {
	ID string
}

type ListVehiclesParam struct {
	DriverID string
}

type UpdateVehicleStatusParam struct {
	ID     string
	Status string
}

type UpdateVehicleStatusResult struct {
	Message string
}
