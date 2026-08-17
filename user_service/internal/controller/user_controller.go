package controller

import (
	"context"
	"user_service/internal/biz"
	"user_service/internal/mapper"
	"user_service/internal/entity"
	userv1 "github.com/logistic/api/logistic/user_service/v1"
	"github.com/google/uuid"
)

type userController struct {
	userv1.UnimplementedUserServiceServer
	engine biz.UserEngine
	mapper mapper.AppMapper
}

func NewUserController(engine biz.UserEngine, appMapper mapper.AppMapper) userv1.UserServiceServer {
	return &userController{
		engine: engine,
		mapper: appMapper,
	}
}

func (c *userController) RegisterUser(ctx context.Context, req *userv1.RegisterUserRequest) (*userv1.RegisterUserResponse, error) {
	param := &entity.RegisterUserParam{
		Phone:    req.Phone,
		Email:    req.Email,
		Password: req.Password,
		Role:     req.Role,
	}

	res, err := c.engine.RegisterUser(ctx, param)
	if err != nil {
		return nil, err
	}

	// Convert string ID to bytes
	parsedID, _ := uuid.Parse(res.ID)

	return &userv1.RegisterUserResponse{
		Id:      parsedID[:],
		Message: res.Message,
	}, nil
}

func (c *userController) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	parsedID, _ := uuid.FromBytes(req.Id)
	param := &entity.GetUserParam{
		ID: parsedID.String(),
	}

	res, err := c.engine.GetUser(ctx, param)
	if err != nil {
		return nil, err
	}

	pbResp := &userv1.GetUserResponse{
		User: c.mapper.EntityUserToPbUser(*res.User),
	}

	if res.DriverProfile != nil {
		pbResp.DriverProfile = c.mapper.EntityDriverProfileToPbDriverProfile(*res.DriverProfile)
	}

	if res.ShipperProfile != nil {
		pbResp.ShipperProfile = c.mapper.EntityShipperProfileToPbShipperProfile(*res.ShipperProfile)
	}

	return pbResp, nil
}

func (c *userController) UpdateDriverKYC(ctx context.Context, req *userv1.UpdateDriverKYCRequest) (*userv1.UpdateDriverKYCResponse, error) {
	parsedID, _ := uuid.FromBytes(req.UserId)
	param := &entity.UpdateDriverKYCParam{
		UserID:    parsedID.String(),
		KycStatus: req.KycStatus,
	}

	res, err := c.engine.UpdateDriverKYC(ctx, param)
	if err != nil {
		return nil, err
	}

	return &userv1.UpdateDriverKYCResponse{
		Message: res.Message,
	}, nil
}
