package biz

import (
	"context"

	"vehicle_service/internal/entity"

	"github.com/google/uuid"
)

type VehicleRepo interface {
	CreateVehicle(ctx context.Context, param *entity.RegisterVehicleParam) (*entity.Vehicle, error)
	GetVehicleByID(ctx context.Context, id uuid.UUID) (*entity.Vehicle, error)
	GetVehiclesByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]entity.Vehicle, error)
	ListVehicles(ctx context.Context, param *entity.ListVehiclesParam) ([]entity.Vehicle, int64, error)
	AdminListVehicles(ctx context.Context, param *entity.AdminListVehiclesParam) ([]entity.Vehicle, int64, error)
	UpdateVehicle(ctx context.Context, param *entity.UpdateVehicleParam) (*entity.Vehicle, error)
	UpdateVehicleStatus(ctx context.Context, id uuid.UUID, status string) (*entity.Vehicle, error)
	UpdateVerification(ctx context.Context, param *entity.VerifyVehicleParam, status string) (*entity.Vehicle, error)
	DeleteVehicle(ctx context.Context, id uuid.UUID) error
	CountVehicles(ctx context.Context, status, verificationStatus string) (int64, error)

	CreateDocument(ctx context.Context, param *entity.UploadDocumentParam) (*entity.VehicleDocument, error)
	GetDocument(ctx context.Context, id uuid.UUID) (*entity.VehicleDocument, error)
	ListDocuments(ctx context.Context, param *entity.ListDocumentsParam) ([]entity.VehicleDocument, error)
	ListPendingDocuments(ctx context.Context, page, pageSize int) ([]entity.VehicleDocument, int64, error)
	ReviewDocument(ctx context.Context, param *entity.ReviewDocumentParam, status string) (*entity.VehicleDocument, error)
	DeleteDocument(ctx context.Context, id uuid.UUID) error
	CountPendingDocuments(ctx context.Context) (int64, error)

	UpsertLocation(ctx context.Context, param *entity.ReportLocationParam, zoneID string) (*entity.VehicleLocation, error)
	GetLocation(ctx context.Context, vehicleID uuid.UUID) (*entity.VehicleLocation, error)

	UpsertAvailability(ctx context.Context, param *entity.SetAvailabilityParam, zoneID string) (*entity.DriverAvailability, error)
	GetAvailability(ctx context.Context, driverID uuid.UUID) (*entity.DriverAvailability, error)
	GetAvailabilitiesByVehicleIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]entity.DriverAvailability, error)
	CountOnlineDrivers(ctx context.Context) (int64, error)

	SearchNearby(ctx context.Context, param *entity.SearchNearbyParam) ([]entity.NearbyVehicle, error)
}