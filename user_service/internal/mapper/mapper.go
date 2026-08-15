package mapper

import (
	"time"

	"github.com/google/uuid"
	pb "github.com/logistic/api/logistic/user_service/v1"

	"user_service/ent"
	"user_service/ent/user"
	"user_service/internal/entity"
)

// goverter:converter
// goverter:useZeroValueOnPointerInconsistency
// goverter:ignoreUnexported
// goverter:extend IdentityTime
//
//go:generate go run github.com/jmattheis/goverter/cmd/goverter@latest gen ./
type AppMapper interface {
	// ==================== REPO MAPPER ====================

	// goverter:map Role Role | EntUserRoleToString
	// goverter:map Status Status | EntUserStatusToString
	EntUserToEntityUser(source *ent.User) entity.User

	// goverter:ignore UserID
	EntDriverProfileToEntityDriverProfile(source *ent.DriverProfile) entity.DriverProfile

	// goverter:ignore UserID
	EntShipperProfileToEntityShipperProfile(source *ent.ShipperProfile) entity.ShipperProfile

	// ==================== CONTROLLER MAPPER ====================

	// goverter:map ID Id | UUIDToString
	// goverter:map CreatedAt CreatedAt | TimeToString
	// goverter:map UpdatedAt UpdatedAt | TimeToString
	EntityUserToPbUser(source entity.User) *pb.User

	// goverter:map Id ID | StringToUUID
	// goverter:ignore CreatedAt
	// goverter:ignore UpdatedAt
	// goverter:ignore PasswordHash
	PbUserToEntityUser(req *pb.User) (entity.User, error)

	// goverter:map UserID UserId | UUIDToString
	// goverter:map IDCard IdCard
	// goverter:map Rating Rating | Float64ToFloat32
	// goverter:map KycStatus KycStatus
	EntityDriverProfileToPbDriverProfile(source entity.DriverProfile) *pb.DriverProfile

	// goverter:map UserID UserId | UUIDToString
	EntityShipperProfileToPbShipperProfile(source entity.ShipperProfile) *pb.ShipperProfile
}

// ==================== HELPERS ====================

func IdentityTime(t time.Time) time.Time {
	return t
}

func EntUserRoleToString(r user.Role) string {
	return string(r)
}

func EntUserStatusToString(s user.Status) string {
	return string(s)
}

func EntDriverProfileUserID(b uuid.UUID) uuid.UUID {
	return b
}

func EntShipperProfileUserID(b uuid.UUID) uuid.UUID {
	return b
}

func StringToUUID(s string) (uuid.UUID, error) {
	if s == "" {
		return uuid.Nil, nil
	}
	return uuid.Parse(s)
}

func UUIDToString(u uuid.UUID) string {
	if u == uuid.Nil {
		return ""
	}
	return u.String()
}

func Float64ToFloat32(f float64) float32 {
	return float32(f)
}

func TimeToString(t time.Time) string {
	return t.Format(time.RFC3339)
}
