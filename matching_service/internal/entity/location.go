package entity

// Location represents a geographic location and logical zone
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	ZoneID    string  `json:"zone_id"` // E.g., "HCM-Q1"
}
