package entity

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Phone        string
	Email        string
	PasswordHash string
	Role         string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type DriverProfile struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	LicenseNumber string
	IDCard        string
	Rating        float64
	KycStatus     string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type ShipperProfile struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	CompanyName string
	TaxCode     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type RegisterUserParam struct {
	Phone    string
	Email    string
	Password string
	Role     string
}

type RegisterUserResult struct {
	ID      string
	Message string
}

type GetUserParam struct {
	ID string
}

type GetUserResult struct {
	User           *User
	DriverProfile  *DriverProfile
	ShipperProfile *ShipperProfile
}

type UpdateDriverKYCParam struct {
	UserID    string
	KycStatus string
}

type UpdateDriverKYCResult struct {
	Message string
}
