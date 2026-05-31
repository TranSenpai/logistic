package biz

import (
	"auth_service/internal/entity"
	"context"
)

type AuthRepo interface {
	FindByEmail(ctx context.Context, email string) (*entity.UserProfile, string, error)

	Save(ctx context.Context, user entity.UserRegister, hashedPassword string) (*entity.UserProfile, error)

	ExistsByEmail(ctx context.Context, email string) (bool, error)
}
