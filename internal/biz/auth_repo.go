package biz

import (
	"context"
	"goBackend/internal/entity"
)

type AuthRepo interface {
	FindByEmail(ctx context.Context, email string) (*entity.UserProfile, string, error)

	Save(ctx context.Context, user entity.UserRegister, hashedPassword string) (*entity.UserProfile, error)

	ExistsByEmail(ctx context.Context, email string) (bool, error)
}

type AuthUsecase interface {
	Register(ctx context.Context, req entity.UserRegister) (*entity.UserProfile, error)

	Login(ctx context.Context, req entity.UserLogin) (*entity.AuthTokenPair, error)
}
