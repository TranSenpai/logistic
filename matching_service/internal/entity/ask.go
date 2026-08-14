package entity

import (
	"time"

	"github.com/google/uuid"
)

type Ask struct {
	ID                uuid.UUID `json:"id"`
	DriverID          uuid.UUID `json:"driver_id"`
	DriverPhone       string    `json:"driver_phone"`
	DriverMail        string    `json:"driver_mail"`
	VehicleID         uuid.UUID `json:"vehicle_id"`
	VehicleType       int8      `json:"vehicle_type"`
	CurrentLocation   Location  `json:"current_location"`
	Destination       Location  `json:"destination"`
	RouteID           []byte    `json:"route_id"`
	CapacityVolumeCbm float64   `json:"capacity_volume_cbm"`
	AvailableVolumeM3 float64   `json:"available_volume_m3"`
	CapacityWeightKg  float64   `json:"capacity_weight_kg"`
	AvailableWeightKg float64   `json:"available_weight_kg"`
	MinPrice          float64   `json:"min_price"`
	DesiredDeposit    float64   `json:"desired_deposit"`
	Status            int8      `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
	ExpiresAt         time.Time `json:"expires_at"`
}

const (
	AskStatusPending int8 = iota
	AskStatusMatched
	AskStatusCancelled
)

func IsValidAskStatus(status int8) bool {
	switch status {
	case AskStatusPending, AskStatusMatched, AskStatusCancelled:
		return true
	default:
		return false
	}
}

func (a *Ask) StatusString() string {
	switch a.Status {
	case AskStatusPending:
		return "PENDING"
	case AskStatusMatched:
		return "MATCHED"
	case AskStatusCancelled:
		return "CANCELLED"
	default:
		return "UNKNOWN"
	}
}
