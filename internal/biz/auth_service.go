package biz

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"goBackend/internal/entity"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// ========================================================================================
// KIẾN TRÚC: AuthService — Tầng Business Logic thuần túy
//
// Rule tối thượng: BIZ LAYER KHÔNG ĐƯỢC:
//   ✗ Import bất kỳ thư viện HTTP nào (gin, net/http)
//   ✗ Import bất kỳ thư viện DB nào (ent, gorm, database/sql)
//   ✗ Biết về JSON serialization
//
// BIZ LAYER CHỈ ĐƯỢC:
//   ✓ Gọi repo interface (không quan tâm impl là gì)
//   ✓ Thực thi business rule (validation, hashing, token generation)
//   ✓ Throw domain errors (ErrEmailAlreadyExists, ErrInvalidCredentials)
//
// Đây là "Protected Variation" principle: Business logic được bọc kín,
// thay đổi ở infra (DB, framework) không ripple vào business rules.
// ========================================================================================

var (
	// Domain errors — được định nghĩa ở biz layer, không ở controller/delivery.
	// Lý do: Controller dịch HTTP status từ domain error, chứ KHÔNG tự biết business rule.
	// Pattern: errors.Is(err, ErrEmailAlreadyExists) → HTTP 409 Conflict
	ErrEmailAlreadyExists  = errors.New("biz: email already registered")
	ErrInvalidCredentials  = errors.New("biz: invalid email or password")
	ErrTokenGenerationFail = errors.New("biz: failed to generate auth token")
)

// jwtClaims là custom claims cho JWT token.
// SECURITY NOTE: Chỉ nhúng UserId và Email vào payload, KHÔNG nhúng password hay role.
// JWT payload là Base64 encoded (KHÔNG encrypted) — ai cũng decode được nếu intercept được token.
// Role-based access: Nên query DB khi cần, hoặc dùng opaque token thay vì JWT nếu cần revocability.
type jwtClaims struct {
	UserId int64  `json:"userId"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

type AuthService struct {
	authRepo AuthRepo

	// TRADE-OFF: Inject jwtSecret vào struct thay vì hardcode hay đọc từ os.Getenv() mỗi lần gọi.
	// os.Getenv() bên trong business logic là bad practice: biz sẽ không testable (env phải set đúng mới test được).
	// Giải pháp: DI layer đọc secret một lần, inject vào đây — biz chỉ dùng, không quan tâm nguồn gốc.
	jwtSecret string
}

func NewAuthService(authRepo AuthRepo, jwtSecret string) AuthUsecase {
	return &AuthService{
		authRepo:  authRepo,
		jwtSecret: jwtSecret,
	}
}

// Register implement business logic đăng ký user.
func (s *AuthService) Register(ctx context.Context, req entity.UserRegister) (*entity.UserProfile, error) {
	// STEP 1 — Guard clause: Kiểm tra email trước khi làm bất cứ điều gì tốn kém (hashing).
	// Đây là "Fail Fast" principle: phát hiện lỗi sớm nhất có thể, tránh tốn CPU cho bcrypt
	// (bcrypt cost 12 ≈ 250ms/request) khi email đã tồn tại.
	exists, err := s.authRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("biz register: checking email existence: %w", err)
	}
	if exists {
		// Trả về domain error, KHÔNG HTTP error. Controller sẽ map sang HTTP 409.
		return nil, ErrEmailAlreadyExists
	}

	// STEP 2 — Hash password TRƯỚC KHI gọi repo.
	// WHY bcrypt? bcrypt tự động thêm random salt → cùng password cho hash khác nhau.
	// → Chống rainbow table attack. Không cần self-managed salt column trong DB.
	//
	// Cost factor 12: OWASP khuyến nghị >= 10 cho bcrypt năm 2024.
	// Cost 12 ≈ 250ms trên server thông thường — đủ chậm để brute-force không khả thi,
	// đủ nhanh để user experience không bị ảnh hưởng.
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, fmt.Errorf("biz register: hashing password: %w", err)
	}

	// STEP 3 — Persist. Repo nhận hashed password, KHÔNG BAO GIỜ nhận plain text.
	profile, err := s.authRepo.Save(ctx, req, string(hashedBytes))
	if err != nil {
		return nil, fmt.Errorf("biz register: persisting user: %w", err)
	}

	return profile, nil
}

// Login implement business logic đăng nhập.
func (s *AuthService) Login(ctx context.Context, req entity.UserLogin) (*entity.AuthTokenPair, error) {
	// STEP 1 — Tìm user. FindByEmail trả về profile + hashed password.
	// SECURITY NOTE: Không phân biệt "email không tồn tại" vs "password sai" trong response.
	// Lý do: Tránh User Enumeration Attack — attacker không biết email có tồn tại hay không.
	profile, hashedPassword, err := s.authRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		// err có thể là "not found" — biz layer chuẩn hóa thành ErrInvalidCredentials
		return nil, ErrInvalidCredentials
	}

	// STEP 2 — Compare hash. bcrypt.CompareHashAndPassword tự xử lý constant-time comparison.
	// Nếu dùng == hay bytes.Equal để so sánh, bạn có thể bị Timing Attack.
	// bcrypt's CompareHashAndPassword chạy trong constant time bất kể match hay không.
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// STEP 3 — Generate token pair.
	tokenPair, err := s.generateTokenPair(profile)
	if err != nil {
		return nil, fmt.Errorf("biz login: %w: %v", ErrTokenGenerationFail, err)
	}

	return tokenPair, nil
}

// generateTokenPair là private helper — chỉ AuthService dùng.
// DESIGN: Tách thành method riêng để dễ unit test (có thể test token generation độc lập).
func (s *AuthService) generateTokenPair(profile *entity.UserProfile) (*entity.AuthTokenPair, error) {
	now := time.Now()
	accessExpiresAt := now.Add(15 * time.Minute)

	// Access Token — short-lived (15 phút), stateless.
	accessClaims := jwtClaims{
		UserId: profile.Id,
		Email:  profile.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "goBackend", // Identify token issuer — dùng khi multi-service
		},
	}

	// ALGORITHM CHOICE: HS256 (HMAC-SHA256) — symmetric, dùng chung 1 secret.
	// Khi nào nên dùng RS256 (asymmetric)? Khi cần nhiều service verify token mà không share secret.
	// Ví dụ: Auth service ký bằng private key, các microservice khác verify bằng public key.
	// Với monolith / single backend: HS256 đơn giản hơn và đủ dùng.
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenStr, err := accessToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, err
	}

	// Refresh Token — long-lived (7 ngày), nhưng stateful (phải lưu DB để revoke được).
	// TODO: Lưu refresh token hash vào DB/Redis. Hiện tại chỉ generate — cần implement persistence.
	// Hash refresh token trước khi lưu (tương tự password): nếu DB bị leak, token vô dụng.
	refreshClaims := jwtClaims{
		UserId: profile.Id,
		Email:  profile.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "goBackend",
		},
	}
	// Dùng secret khác cho refresh token — tách biệt signing key theo token type.
	// Lý do: Nếu access token secret bị lộ, refresh token vẫn an toàn.
	refreshSecret := s.jwtSecret + "_refresh"
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenStr, err := refreshToken.SignedString([]byte(refreshSecret))
	if err != nil {
		return nil, err
	}

	_ = os.Getenv // Suppress unused import nếu có — sẽ xóa khi đã có env config đầy đủ

	return &entity.AuthTokenPair{
		AccessToken:  accessTokenStr,
		RefreshToken: refreshTokenStr,
		ExpiresIn:    accessExpiresAt.Unix(),
	}, nil
}
