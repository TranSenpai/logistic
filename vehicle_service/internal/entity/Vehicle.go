package entity

import "time"

type VehicleType int8

const (
	Motorbike VehicleType = iota
	Car
	Van
	Truck
	Container
)

type VehicleBrand int8

const (
	Toyota VehicleBrand = iota
	Thaco
	Kimlong
	Huyndai
	Mitsubisi
)

type Vehicle struct {
	Id             []byte
	Type           VehicleType
	Brand          VehicleBrand
	InspectionDate time.Time
}
