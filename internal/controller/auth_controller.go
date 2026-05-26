package controller

import (
	"errors"
	"net/http"
	"time"

	"goBackend/internal/biz"
	authdto "goBackend/internal/dto/gen"
	"goBackend/internal/entity"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authUsecase biz.AuthUsecase
}

func NewAuthController(authUsecase biz.AuthUsecase) *AuthController {
	return &AuthController{authUsecase: authUsecase}
}

func (a *AuthController) Register(ctx *gin.Context) {
	var req authdto.RegisterRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_failed",
			"message": err.Error(),
		})
		return
	}

	domainReq := entity.UserRegister{
		Email:    req.Email,
		FullName: req.FullName,
		Password: req.Password,
	}

	profile, err := a.authUsecase.Register(ctx.Request.Context(), domainReq)
	if err != nil {
		switch {
		case errors.Is(err, biz.ErrEmailAlreadyExists):
			ctx.JSON(http.StatusConflict, gin.H{
				"error":   "email_already_exists",
				"message": "Email đã được đăng ký. Vui lòng dùng email khác hoặc đăng nhập.",
			})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_server_error",
				"message": "Đã xảy ra lỗi. Vui lòng thử lại sau.",
			})
		}
		return
	}

	resp := buildAuthResponse(profile, &entity.AuthTokenPair{})

	ctx.JSON(http.StatusCreated, resp)
}

func (a *AuthController) Login(ctx *gin.Context) {
	var req authdto.LoginRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_failed",
			"message": err.Error(),
		})
		return
	}

	tokenPair, err := a.authUsecase.Login(ctx.Request.Context(), entity.UserLogin{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, biz.ErrInvalidCredentials):
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"error":   "invalid_credentials",
				"message": "Email hoặc mật khẩu không đúng.",
			})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_server_error",
				"message": "Đã xảy ra lỗi. Vui lòng thử lại sau.",
			})
		}
		return
	}

	ctx.JSON(http.StatusOK, authdto.AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	})
}

func buildAuthResponse(profile *entity.UserProfile, tokens *entity.AuthTokenPair) authdto.AuthResponse {
	userResp := authdto.UserProfileResponse{
		Id:       profile.Id,
		Email:    profile.Email,
		FullName: profile.FullName,
		Avatar:   profile.Avatar,
	}

	if profile.CreatedAt != nil {
		ts := profile.CreatedAt.Unix()
		userResp.CreatedAt = &ts
		_ = time.Now
	}

	return authdto.AuthResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
		User:         userResp,
	}
}
