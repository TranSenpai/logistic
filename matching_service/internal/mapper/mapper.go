package mapper

import (
	"matching_service/ent"
	"matching_service/internal/entity"
	"strconv"
	"strings"
	"time"
)

// goverter:converter
// goverter:useZeroValueOnPointerInconsistency
// goverter:ignoreUnexported
// goverter:extend IdentityTime
//
//go:generate go run github.com/jmattheis/goverter/cmd/goverter@latest gen ./
type Converter interface {
	// goverter:map ID | IntToString
	// goverter:map DriverID | IntToString
	// goverter:map CurrentCoordinates CurrentLocation | ParseLocation
	// goverter:map Status | AskStatusToString
	// goverter:map MinPrice | Float64PtrToFloat64
	// goverter:ignore VehicleID
	// goverter:ignore Destination
	// goverter:ignore ExpiresAt
	EntAskToEntityAsk(source *ent.Ask) entity.Ask
	EntAskListToEntityAskList(source []*ent.Ask) []entity.Ask
	// goverter:map ID | IntToString
	// goverter:map UserID | IntToString
	// goverter:map PickupCoordinates Origin | ParseLocation
	// goverter:map DeliveryCoordinates Destination | ParseLocation
	// goverter:map Status | BidStatusToString
	// goverter:map MaxPrice | Float64PtrToFloat64
	// goverter:ignore ExpiresAt
	EntBidToEntityBid(source *ent.Bid) entity.Bid
	EntBidListToEntityBidList(source []*ent.Bid) []entity.Bid
}

func IntToString(i int) string {
	return strconv.Itoa(i)
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

func ParseLocation(pointStr string) entity.Location {
	pointStr = strings.TrimPrefix(pointStr, "POINT(")
	pointStr = strings.TrimSuffix(pointStr, ")")
	parts := strings.Split(pointStr, " ")
	if len(parts) != 2 {
		return entity.Location{}
	}
	lng, _ := strconv.ParseFloat(parts[0], 64)
	lat, _ := strconv.ParseFloat(parts[1], 64)
	return entity.Location{
		Longitude: lng,
		Latitude:  lat,
	}
}

func AskStatusToString(status int) string {
	switch status {
	case 1:
		return entity.AskStatusPending
	case 2:
		return entity.AskStatusMatched
	case 3:
		return entity.AskStatusCancelled
	default:
		return "UNKNOWN"
	}
}

func BidStatusToString(status int) string {
	switch status {
	case 1:
		return entity.BidStatusPending
	case 2:
		return entity.BidStatusMatched
	case 3:
		return entity.BidStatusCancelled
	default:
		return "UNKNOWN"
	}
}

func BidStatusToInt(status string) int {
	switch status {
	case entity.BidStatusPending:
		return 1
	case entity.BidStatusMatched:
		return 2
	case entity.BidStatusCancelled:
		return 3
	default:
		return -1
	}
}

func AskStatusToInt(status string) int {
	switch status {
	case entity.AskStatusPending:
		return 1
	case entity.AskStatusMatched:
		return 2
	case entity.AskStatusCancelled:
		return 3
	default:
		return -1
	}
}
