package delivery

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	authdto "goBackend/api/logistics/v1/gen"
	"goBackend/auth_service/internal/biz"
	"goBackend/auth_service/internal/entity"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *HttpHandler) Register(ctx *gin.Context) {
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

	profile, err := h.authUsecase.Register(ctx.Request.Context(), domainReq)
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

	// Dùng Goverter để map thay vì hàm thủ công
	userResp := h.authMapper.ToUserProfileResponse(profile)
	resp := authdto.AuthResponse{
		User: userResp,
	}

	ctx.JSON(http.StatusCreated, resp)
}

func (h *HttpHandler) Login(ctx *gin.Context) {
	var req authdto.LoginRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_failed",
			"message": err.Error(),
		})
		return
	}

	tokenPair, err := h.authUsecase.Login(ctx.Request.Context(), entity.UserLogin{
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

func (h *HttpHandler) GoogleLogin(ctx *gin.Context) {
	// 1. Tạo state ngẫu nhiên
	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)

	// 2. Lấy URL từ service
	url := h.authUsecase.GetGoogleLoginURL(state)

	// 3. Set cookie và redirect
	ctx.SetCookie("oauth_state", state, int(time.Minute*5), "/", "", false, true)
	ctx.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *HttpHandler) GoogleCallback(ctx *gin.Context) {
	urlState := ctx.Query("state")
	cookieState, err := ctx.Cookie("oauth_state")
	if err != nil || urlState != cookieState {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_state",
			"message": "State không hợp lệ.",
		})
		return
	}
	code := ctx.Query("code")
	tokenPair, err := h.authUsecase.GoogleCallback(ctx.Request.Context(), code)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_server_error",
			"message": "Đã xảy ra lỗi. Vui lòng thử lại sau.",
		})
		return
	}

	// Tính toán thời gian sống của cookie (MaxAge tính bằng giây)
	accessMaxAge := max(int(tokenPair.ExpiresIn-time.Now().Unix()), 0)
	refreshMaxAge := int(7 * 24 * time.Hour / time.Second)

	// Set cookie (Domain để trống hoặc tuỳ chỉnh theo frontend domain, HttpOnly = false cho access token để FE có thể đọc nếu cần)
	ctx.SetCookie("access_token", tokenPair.AccessToken, accessMaxAge, "/", "", false, false)
	ctx.SetCookie("refresh_token", tokenPair.RefreshToken, refreshMaxAge, "/", "", false, true) // Refresh token nên set HttpOnly=true

	// Redirect user về trang chủ của Frontend (Dùng 127.0.0.1 để khớp domain với Backend)
	ctx.Redirect(http.StatusTemporaryRedirect, "http://127.0.0.1:3000/")
}

func (h *HttpHandler) GetInfo(ctx *gin.Context) {
	// Lấy token từ cookie
	token, err := ctx.Cookie("access_token")
	if err != nil {
		// Thử lấy từ header Authorization: Bearer <token>
		authHeader := ctx.GetHeader("Authorization")
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token = authHeader[7:]
		}
	}

	if token == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Không tìm thấy token.",
		})
		return
	}

	// Validate token và lấy profile
	profile, err := h.authUsecase.VerifyToken(ctx.Request.Context(), token)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error":   "invalid_token",
			"message": "Token không hợp lệ hoặc đã hết hạn.",
		})
		return
	}

	// Trả về JSON theo đúng chuẩn TRes của Frontend
	userResp := h.authMapper.ToUserProfileResponse(profile)
	ctx.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"statusCode": 200,
		"message":    "Get info successfully",
		"data": gin.H{
			"user":   userResp,
			"isTotp": false,
		},
	})
}
