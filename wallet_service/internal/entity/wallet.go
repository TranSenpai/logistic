package entity

import (
	"time"

	"github.com/google/uuid"
)

// Wallet represents a core business entity, independent of any ORM or database framework.
type Wallet struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	UserType  uint8
	Balance   int64
	Currency  string
	Status    uint8
	CreatedAt time.Time
	UpdatedAt time.Time
}
