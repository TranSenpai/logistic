package handler

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"

	pb "github.com/logistic/api/logistic/auth_service/v1"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthHandler struct {
	authClient pb.AuthServiceClient
}

func NewAuthHandler(authClient pb.AuthServiceClient) *AuthHandler {
	return &AuthHandler{
		authClient: authClient,
	}
}

// Structs for parsing HTTP requests
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	FullName string `json:"full_name"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Register(ctx *gin.Context) {
	var req RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "validation_failed", "message": err.Error()})
		return
	}

	resp, err := h.authClient.Register(ctx.Request.Context(), &pb.RegisterRequest{
		Email:    req.Email,
		FullName: req.FullName,
		Password: req.Password,
	})
	if err != nil {
		st, _ := status.FromError(err)
		if st.Code() == codes.AlreadyExists {
			ctx.JSON(http.StatusConflict, gin.H{"error": "email_already_exists", "message": st.Message()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal_server_error", "message": st.Message()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"user": gin.H{
			"id":        resp.Profile.Id,
			"email":     resp.Profile.Email,
			"full_name": resp.Profile.FullName,
			"avatar":    resp.Profile.Avatar,
		},
	})
}

func (h *AuthHandler) Login(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "validation_failed", "message": err.Error()})
		return
	}

	resp, err := h.authClient.Login(ctx.Request.Context(), &pb.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		st, _ := status.FromError(err)
		if st.Code() == codes.Unauthenticated {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_credentials", "message": st.Message()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal_server_error", "message": st.Message()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"access_token":  resp.TokenPair.AccessToken,
		"refresh_token": resp.TokenPair.RefreshToken,
		"expires_in":    resp.TokenPair.ExpiresIn,
	})
}

func (h *AuthHandler) GoogleLogin(ctx *gin.Context) {
	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)

	resp, err := h.authClient.GetGoogleLoginURL(ctx.Request.Context(), &pb.GetGoogleLoginURLRequest{
		State: state,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal_server_error"})
		return
	}

	ctx.SetCookie("oauth_state", state, int(time.Minute*5), "/", "", false, true)
	ctx.Redirect(http.StatusTemporaryRedirect, resp.Url)
}

func (h *AuthHandler) GoogleCallback(ctx *gin.Context) {
	urlState := ctx.Query("state")
	cookieState, err := ctx.Cookie("oauth_state")
	if err != nil || urlState != cookieState {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid_state", "message": "State không hợp lệ."})
		return
	}
	
	code := ctx.Query("code")
	resp, err := h.authClient.GoogleCallback(ctx.Request.Context(), &pb.GoogleCallbackRequest{
		Code: code,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal_server_error"})
		return
	}

	tokenPair := resp.TokenPair
	accessMaxAge := max(int(tokenPair.ExpiresIn-time.Now().Unix()), 0)
	refreshMaxAge := int(7 * 24 * time.Hour / time.Second)

	ctx.SetCookie("access_token", tokenPair.AccessToken, accessMaxAge, "/", "", false, false)
	ctx.SetCookie("refresh_token", tokenPair.RefreshToken, refreshMaxAge, "/", "", false, true)

	ctx.Redirect(http.StatusTemporaryRedirect, "http://127.0.0.1:3000/")
}

func (h *AuthHandler) GetInfo(ctx *gin.Context) {
	token, err := ctx.Cookie("access_token")
	if err != nil {
		authHeader := ctx.GetHeader("Authorization")
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token = authHeader[7:]
		}
	}

	if token == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Không tìm thấy token."})
		return
	}

	resp, err := h.authClient.VerifyToken(ctx.Request.Context(), &pb.VerifyTokenRequest{
		Token: token,
	})
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_token", "message": "Token không hợp lệ hoặc đã hết hạn."})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"statusCode": 200,
		"message":    "Get info successfully",
		"data": gin.H{
			"user": gin.H{
				"id":        resp.Profile.Id,
				"email":     resp.Profile.Email,
				"full_name": resp.Profile.FullName,
				"avatar":    resp.Profile.Avatar,
			},
			"isTotp": false,
		},
	})
}
