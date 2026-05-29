package entity

import "time"

// Location represents a geographic location and logical zone
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	ZoneID    string  `json:"zone_id"` // E.g., "HCM-Q1"
}

// Bid represents a cargo waiting for a vehicle (Demand)
type Bid struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Origin      Location  `json:"origin"`
	Destination Location  `json:"destination"`
	VolumeM3    float64   `json:"volume_m3"`
	WeightKg    float64   `json:"weight_kg"`
	MaxPrice    float64   `json:"max_price"`
	Status      string    `json:"status"` // PENDING, MATCHED, EXPIRED
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
}

// Ask represents an empty vehicle waiting for cargo (Supply)
type Ask struct {
	ID                string    `json:"id"`
	VehicleID         string    `json:"vehicle_id"`
	DriverID          string    `json:"driver_id"`
	CurrentLocation   Location  `json:"current_location"`
	Destination       Location  `json:"destination"`
	AvailableVolumeM3 float64   `json:"available_volume_m3"`
	AvailableWeightKg float64   `json:"available_weight_kg"`
	MinPrice          float64   `json:"min_price"`
	Status            string    `json:"status"` // PENDING, MATCHED, EXPIRED
	ExpiresAt         time.Time `json:"expires_at"`
	CreatedAt         time.Time `json:"created_at"`
}

// MatchResult represents a successful match between a Bid and an Ask
type MatchResult struct {
	BidID     string    `json:"bid_id"`
	AskID     string    `json:"ask_id"`
	Price     float64   `json:"price"`
	MatchedAt time.Time `json:"matched_at"`
}
