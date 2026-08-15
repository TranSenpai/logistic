package biz

import (
	"context"
	"fmt"

	"user_service/ent"
	"user_service/ent/user"
	"user_service/internal/entity"

	"github.com/google/uuid"
)

type UserEngine interface {
	RegisterUser(ctx context.Context, param *entity.RegisterUserParam) (*entity.RegisterUserResult, error)
	GetUser(ctx context.Context, param *entity.GetUserParam) (*entity.GetUserResult, error)
	UpdateDriverKYC(ctx context.Context, param *entity.UpdateDriverKYCParam) (*entity.UpdateDriverKYCResult, error)
}

type userEngineImpl struct {
	repo UserRepo
}

func NewUserEngine(repo UserRepo) UserEngine {
	return &userEngineImpl{repo: repo}
}

func (e *userEngineImpl) RegisterUser(ctx context.Context, param *entity.RegisterUserParam) (*entity.RegisterUserResult, error) {
	if _, err := e.repo.GetUserByPhone(ctx, param.Phone); err == nil {
		return nil, fmt.Errorf("phone number already registered")
	}

	u := &ent.User{
		Phone:        param.Phone,
		Email:        param.Email,
		PasswordHash: param.Password, 
		Role:         user.Role(param.Role),
	}

	createdUser, err := e.repo.CreateUser(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	if param.Role == "driver" {
		dp := &ent.DriverProfile{}
		dp.Edges.User = createdUser
		if err := e.repo.CreateDriverProfile(ctx, dp); err != nil {
			return nil, fmt.Errorf("failed to create driver profile: %w", err)
		}
	} else if param.Role == "shipper" {
		sp := &ent.ShipperProfile{}
		sp.Edges.User = createdUser
		if err := e.repo.CreateShipperProfile(ctx, sp); err != nil {
			return nil, fmt.Errorf("failed to create shipper profile: %w", err)
		}
	}

	return &entity.RegisterUserResult{
		ID:      createdUser.ID.String(),
		Message: "User registered successfully",
	}, nil
}

func (e *userEngineImpl) GetUser(ctx context.Context, param *entity.GetUserParam) (*entity.GetUserResult, error) {
	uid, err := uuid.Parse(param.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID")
	}

	u, err := e.repo.GetUserByID(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	result := &entity.GetUserResult{
		User: &entity.User{
			ID:        u.ID,
			Phone:     u.Phone,
			Email:     u.Email,
			Role:      string(u.Role),
			Status:    string(u.Status),
			CreatedAt: u.CreatedAt,
			UpdatedAt: u.UpdatedAt,
		},
	}

	if string(u.Role) == "driver" {
		dp, err := e.repo.GetDriverProfile(ctx, uid)
		if err == nil && dp != nil {
			result.DriverProfile = &entity.DriverProfile{
				UserID:        uid,
				LicenseNumber: dp.LicenseNumber,
				IDCard:        dp.IDCard,
				Rating:        dp.Rating,
				KycStatus:     string(dp.KycStatus),
			}
		}
	} else if string(u.Role) == "shipper" {
		sp, err := e.repo.GetShipperProfile(ctx, uid)
		if err == nil && sp != nil {
			result.ShipperProfile = &entity.ShipperProfile{
				UserID:      uid,
				CompanyName: sp.CompanyName,
				TaxCode:     sp.TaxCode,
			}
		}
	}

	return result, nil
}

func (e *userEngineImpl) UpdateDriverKYC(ctx context.Context, param *entity.UpdateDriverKYCParam) (*entity.UpdateDriverKYCResult, error) {
	uid, err := uuid.Parse(param.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID")
	}

	if err := e.repo.UpdateDriverKYC(ctx, uid, param.KycStatus); err != nil {
		return nil, fmt.Errorf("failed to update KYC status: %w", err)
	}

	return &entity.UpdateDriverKYCResult{
		Message: "KYC status updated successfully",
	}, nil
}
