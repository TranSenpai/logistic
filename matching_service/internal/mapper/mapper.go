package mapper

import (
	"matching_service/ent"
	"matching_service/internal/entity"
	"time"
)

// goverter:converter
// goverter:useZeroValueOnPointerInconsistency
// goverter:ignoreUnexported
// goverter:extend IdentityTime
//
//go:generate go run github.com/jmattheis/goverter/cmd/goverter@latest gen ./
type Converter interface {
	// goverter:map . CurrentLocation | MapAskCurrentLocation
	// goverter:map . Destination | MapAskDestination
	// goverter:map MinPrice | Float64PtrToFloat64
	// goverter:ignore ExpiresAt
	EntAskToEntityAsk(source *ent.Asks) entity.Ask
	EntAskListToEntityAskList(source []*ent.Asks) []entity.Ask

	// goverter:map . Origin | MapBidOrigin
	// goverter:map . Destination | MapBidDestination
	// goverter:map MaxPrice | Float64PtrToFloat64
	// goverter:ignore ExpiresAt
	EntBidToEntityBid(source *ent.Bids) entity.Bid
	EntBidListToEntityBidList(source []*ent.Bids) []entity.Bid
}

func Float64PtrToFloat64(f *float64) float64 {
	if f == nil {
		return 0
	}
	return *f
}

func IdentityTime(t time.Time) time.Time {
	return t
}

func MapAskCurrentLocation(source *ent.Asks) entity.Location {
	if source == nil {
		return entity.Location{}
	}

	return entity.Location{
		Latitude:  source.OriginLat,
		Longitude: source.OriginLng,
		ZoneID:    source.ZoneID,
	}
}

func MapAskDestination(source *ent.Asks) entity.Location {
	if source == nil {
		return entity.Location{}
	}

	return entity.Location{
		Latitude:  source.DestinationLat,
		Longitude: source.DestinationLng,
	}
}

func MapBidOrigin(source *ent.Bids) entity.Location {
	if source == nil {
		return entity.Location{}
	}

	return entity.Location{
		Latitude:  source.OriginLat,
		Longitude: source.OriginLng,
		ZoneID:    source.ZoneID,
	}
}

func MapBidDestination(source *ent.Bids) entity.Location {
	if source == nil {
		return entity.Location{}
	}

	return entity.Location{
		Latitude:  source.DestinationLat,
		Longitude: source.DestinationLng,
	}
}
