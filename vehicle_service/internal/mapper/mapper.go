package mapper

import (
	"time"

	"github.com/google/uuid"
	pb "github.com/logistic/api/logistic/vehicle_service/v1"

	"vehicle_service/ent"
	"vehicle_service/ent/vehicle"
	"vehicle_service/internal/entity"
)

// goverter:converter
// goverter:useZeroValueOnPointerInconsistency
// goverter:ignoreUnexported
// goverter:extend IdentityTime
//
//go:generate go run github.com/jmattheis/goverter/cmd/goverter@latest gen ./
type AppMapper interface {
	// ==================== REPO MAPPER ====================

	// goverter:map VehicleType VehicleType | EntVehicleTypeToString
	EntVehicleToEntityVehicle(source *ent.Vehicle) entity.Vehicle
	EntVehicleListToEntityVehicleList(source []*ent.Vehicle) []entity.Vehicle

	// ==================== CONTROLLER MAPPER ====================

	// goverter:map Id ID | StringToUUID
	// goverter:map DriverId DriverID | StringToUUID
	// goverter:map CapacityWeightKg CapacityWeightKg | Float32ToFloat64
	// goverter:map CapacityVolumeCbm CapacityVolumeCbm | Float32ToFloat64
	// goverter:ignore CreatedAt
	// goverter:ignore UpdatedAt
	PbVehicleToEntity(req *pb.Vehicle) (entity.Vehicle, error)

	// goverter:map ID Id | UUIDToString
	// goverter:map DriverID DriverId | UUIDToString
	// goverter:map CapacityWeightKg CapacityWeightKg | Float64ToFloat32
	// goverter:map CapacityVolumeCbm CapacityVolumeCbm | Float64ToFloat32
	EntityVehicleToPbVehicle(source entity.Vehicle) *pb.Vehicle
	EntityVehicleListToPbVehicleList(source []entity.Vehicle) []*pb.Vehicle
}

// ==================== HELPERS ====================

func IdentityTime(t time.Time) time.Time {
	return t
}

func EntVehicleTypeToString(t vehicle.VehicleType) string {
	return string(t)
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

func Float32ToFloat64(f float32) float64 {
	return float64(f)
}

func Float64ToFloat32(f float64) float32 {
	return float32(f)
}
