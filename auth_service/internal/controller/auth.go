package controller

import (
	"context"

	"auth_service/internal/biz"
	"auth_service/internal/entity"
	"auth_service/internal/mapper"

	pb "github.com/logistic/api/logistic/auth_service/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type authController struct {
	pb.UnimplementedAuthServiceServer
	authBiz biz.AuthService
	mapper  mapper.AuthMapper
}

func NewAuthController(authBiz biz.AuthService, mapper mapper.AuthMapper) pb.AuthServiceServer {
	return &authController{
		authBiz: authBiz,
		mapper:  mapper,
	}
}

func (h *authController) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}

	profile, err := h.authBiz.Register(ctx, entity.UserRegister{
		Email:    req.Email,
		FullName: req.FullName,
		Password: req.Password,
		GoogleID: req.GoogleId,
	})
	if err != nil {
		if err == biz.ErrEmailAlreadyExists {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.RegisterResponse{
		Profile: h.mapper.ToUserProfileProto(profile),
	}, nil
}

func (h *authController) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}

	tokenPair, err := h.authBiz.Login(ctx, entity.UserLogin{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if err == biz.ErrInvalidCredentials {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.LoginResponse{
		TokenPair: h.mapper.ToAuthTokenPairProto(tokenPair),
	}, nil
}

func (h *authController) GetGoogleLoginURL(ctx context.Context, req *pb.GetGoogleLoginURLRequest) (*pb.GetGoogleLoginURLResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}

	url := h.authBiz.GetGoogleLoginURL(req.State)
	return &pb.GetGoogleLoginURLResponse{
		Url: url,
	}, nil
}

func (h *authController) GoogleCallback(ctx context.Context, req *pb.GoogleCallbackRequest) (*pb.GoogleCallbackResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}

	tokenPair, err := h.authBiz.GoogleCallback(ctx, req.Code)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GoogleCallbackResponse{
		TokenPair: h.mapper.ToAuthTokenPairProto(tokenPair),
	}, nil
}

func (h *authController) VerifyToken(ctx context.Context, req *pb.VerifyTokenRequest) (*pb.VerifyTokenResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}

	profile, err := h.authBiz.VerifyToken(ctx, req.Token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	return &pb.VerifyTokenResponse{
		Profile: h.mapper.ToUserProfileProto(profile),
	}, nil
}
