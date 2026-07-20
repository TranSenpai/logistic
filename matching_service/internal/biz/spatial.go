package biz

import "context"

type SpatialEngine interface {
	GetZoneId(ctx context.Context, lat, lng float64) (string, error)
	GetNeighborZones(ctx context.Context, zoneID string) ([]string, error)
}
