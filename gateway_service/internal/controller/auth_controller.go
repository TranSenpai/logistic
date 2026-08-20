package controller

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"

	"gateway_service/internal/response"

	pb "github.com/logistic/api/logistic/auth_service/v1"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authClient pb.AuthServiceClient
}

func NewAuthController(authClient pb.AuthServiceClient) *AuthController {
	return &AuthController{
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

// Register godoc
// @Summary      Đăng ký tài khoản
// @Description  Tạo tài khoản mới bằng email và mật khẩu. Gateway chuyển tiếp request tới auth_service qua gRPC.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body RegisterRequest true "Thông tin đăng ký"
// @Success      201 {object} map[string]interface{} "Tạo tài khoản thành công"
// @Failure      400 {object} map[string]interface{} "Lỗi dữ liệu đầu vào"
// @Failure      409 {object} map[string]interface{} "Email đã tồn tại"
// @Failure      500 {object} map[string]interface{} "Lỗi server nội bộ"
// @Router       /api/v1/auth/register [post]
func (c *AuthController) Register(ctx *gin.Context) {
	var req RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "VALIDATION_FAILED", err.Error())
		return
	}

	resp, err := c.authClient.Register(ctx.Request.Context(), &pb.RegisterRequest{
		Email:    req.Email,
		FullName: req.FullName,
		Password: req.Password,
	})
	if err != nil {
		// response.Error dịch codes.AlreadyExists sang HTTP 409 sẵn.
		response.Error(ctx, err)
		return
	}

	response.Created(ctx, gin.H{
		"user": gin.H{
			"id":        resp.Profile.Id,
			"email":     resp.Profile.Email,
			"full_name": resp.Profile.FullName,
			"avatar":    resp.Profile.Avatar,
		},
	}, "Đăng ký tài khoản thành công")
}

// Login godoc
// @Summary      Đăng nhập
// @Description  Đăng nhập bằng email và mật khẩu. Trả về access_token, refresh_token và expires_in.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body LoginRequest true "Thông tin đăng nhập"
// @Success      200 {object} map[string]interface{} "Đăng nhập thành công, trả về token pair"
// @Failure      400 {object} map[string]interface{} "Lỗi dữ liệu đầu vào"
// @Failure      401 {object} map[string]interface{} "Sai email hoặc mật khẩu"
// @Failure      500 {object} map[string]interface{} "Lỗi server nội bộ"
// @Router       /api/v1/auth/login [post]
func (c *AuthController) Login(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "VALIDATION_FAILED", err.Error())
		return
	}

	resp, err := c.authClient.Login(ctx.Request.Context(), &pb.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		// response.Error dịch codes.Unauthenticated sang HTTP 401 sẵn.
		response.Error(ctx, err)
		return
	}

	response.OK(ctx, gin.H{
		"access_token":  resp.TokenPair.AccessToken,
		"refresh_token": resp.TokenPair.RefreshToken,
		"expires_in":    resp.TokenPair.ExpiresIn,
	})
}

// GoogleLogin godoc
// @Summary      Đăng nhập bằng Google OAuth2
// @Description  Khởi tạo luồng OAuth2 với Google. Redirect người dùng đến trang đăng nhập Google. Tạo state cookie để chống CSRF.
// @Tags         Auth - OAuth2
// @Produce      json
// @Success      307 {string} string "Redirect đến Google OAuth consent screen"
// @Failure      500 {object} map[string]interface{} "Lỗi khi lấy Google Login URL"
// @Router       /api/v1/auth/google/login [get]
func (c *AuthController) GoogleLogin(ctx *gin.Context) {
	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)

	resp, err := c.authClient.GetGoogleLoginURL(ctx.Request.Context(), &pb.GetGoogleLoginURLRequest{
		State: state,
	})
	if err != nil {
		response.Error(ctx, err)
		return
	}

	ctx.SetCookie("oauth_state", state, int(time.Minute*5), "/", "", false, true)
	ctx.Redirect(http.StatusTemporaryRedirect, resp.Url)
}

// GoogleCallback godoc
// @Summary      Google OAuth2 Callback
// @Description  Xử lý callback từ Google sau khi user đăng nhập. Xác thực state, đổi code lấy token, set cookie access_token và refresh_token, rồi redirect về frontend.
// @Tags         Auth - OAuth2
// @Produce      json
// @Param        state query string true "OAuth state parameter (CSRF protection)"
// @Param        code  query string true "Authorization code từ Google"
// @Success      307 {string} string "Redirect về frontend với cookie đã set"
// @Failure      400 {object} map[string]interface{} "State không hợp lệ"
// @Failure      500 {object} map[string]interface{} "Lỗi xử lý callback"
// @Router       /api/v1/auth/google/callback [get]
func (c *AuthController) GoogleCallback(ctx *gin.Context) {
	urlState := ctx.Query("state")
	cookieState, err := ctx.Cookie("oauth_state")
	if err != nil || urlState != cookieState {
		response.BadRequest(ctx, "INVALID_OAUTH_STATE", "State không hợp lệ.")
		return
	}

	code := ctx.Query("code")
	resp, err := c.authClient.GoogleCallback(ctx.Request.Context(), &pb.GoogleCallbackRequest{
		Code: code,
	})
	if err != nil {
		response.Error(ctx, err)
		return
	}

	tokenPair := resp.TokenPair
	accessMaxAge := max(int(tokenPair.ExpiresIn-time.Now().Unix()), 0)
	refreshMaxAge := int(7 * 24 * time.Hour / time.Second)

	ctx.SetCookie("access_token", tokenPair.AccessToken, accessMaxAge, "/", "", false, false)
	ctx.SetCookie("refresh_token", tokenPair.RefreshToken, refreshMaxAge, "/", "", false, true)

	ctx.Redirect(http.StatusTemporaryRedirect, "http://127.0.0.1:3000/")
}

// GetInfo godoc
// @Summary      Lấy thông tin người dùng
// @Description  Lấy thông tin profile của người dùng hiện tại. Token có thể truyền qua cookie access_token hoặc header Authorization Bearer.
// @Tags         Auth
// @Produce      json
// @Param        Authorization header string false "Bearer token (nếu không dùng cookie)"
// @Success      200 {object} map[string]interface{} "Lấy thông tin thành công"
// @Failure      401 {object} map[string]interface{} "Token không hợp lệ hoặc không tìm thấy"
// @Router       /api/v1/auth/me [get]
func (c *AuthController) GetInfo(ctx *gin.Context) {
	token, err := ctx.Cookie("access_token")
	if err != nil {
		authHeader := ctx.GetHeader("Authorization")
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token = authHeader[7:]
		}
	}

	if token == "" {
		response.Unauthorized(ctx, "Không tìm thấy token.")
		return
	}

	resp, err := c.authClient.VerifyToken(ctx.Request.Context(), &pb.VerifyTokenRequest{
		Token: token,
	})
	if err != nil {
		response.Unauthorized(ctx, "Token không hợp lệ hoặc đã hết hạn.")
		return
	}

	response.OKMessage(ctx, gin.H{
		"user": gin.H{
			"id":        resp.Profile.Id,
			"email":     resp.Profile.Email,
			"full_name": resp.Profile.FullName,
			"avatar":    resp.Profile.Avatar,
		},
		"isTotp": false,
	}, "Get info successfully")
}
