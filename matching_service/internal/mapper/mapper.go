package mapper

import (
	"time"

	"github.com/google/uuid"
	pb "github.com/logistic/api/logistic/matching_service/v1"
	"google.golang.org/protobuf/types/known/timestamppb"

	"matching_service/ent"
	"matching_service/internal/entity"
)

// goverter:converter
// goverter:useZeroValueOnPointerInconsistency
// goverter:ignoreUnexported
// goverter:extend IdentityTime
//
//go:generate go run github.com/jmattheis/goverter/cmd/goverter@latest gen ./
type AppMapper interface {
	// ==================== REPO MAPPER ====================

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

	// ==================== CONTROLLER MAPPER ====================

	// goverter:ignore ID
	// goverter:ignore CreatedAt
	// goverter:ignore Status
	// goverter:map UserId UserID | BytesToUUID
	// goverter:map ExpiresAt ExpiresAt | TimestampToTime
	SubmitBidReqToEntity(req *pb.SubmitBidRequest) (entity.Bid, error)

	// goverter:ignore ID
	// goverter:ignore CreatedAt
	// goverter:ignore Status
	// goverter:map VehicleId VehicleID | BytesToUUID
	// goverter:map DriverId DriverID | BytesToUUID
	// goverter:map ExpiresAt ExpiresAt | TimestampToTime
	SubmitAskReqToEntity(req *pb.SubmitAskRequest) (entity.Ask, error)

	// goverter:map ZoneId ZoneID
	MapLocation(loc *pb.Location) entity.Location
}

// ==================== HELPERS ====================

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

func BytesToUUID(b []byte) (uuid.UUID, error) {
	if len(b) == 0 {
		return uuid.Nil, nil
	}
	return uuid.FromBytes(b)
}

func TimestampToTime(ts *timestamppb.Timestamp) time.Time {
	if ts != nil && ts.IsValid() {
		return ts.AsTime()
	}
	return time.Time{}
}
