package dto

// ========================================================================================
// DTO (Data Transfer Object) — Tại sao cần tầng này?
//
// Phân biệt ba tầng dữ liệu trong hệ thống này:
//
// 1. DTO (tầng này)    : Giao tiếp với CLIENT (HTTP request/response JSON)
// 2. Entity            : Giao tiếp INTERNAL giữa biz ↔ repo
// 3. Ent Schema        : Giao tiếp với DATABASE
//
// Ví dụ tại sao cần tách:
// - Client gửi lên: { "email": "...", "password": "...", "confirmPassword": "..." }
// - Entity chỉ cần:  { Email, Password }  (confirmPassword là UI concern, không phải domain)
// - DB lưu:          { email, password_hash, ... }
//
// Nếu dùng entity làm DTO: bạn expose internal field names ra API contract.
// Khi refactor nội bộ → API contract thay đổi → breaking change cho client.
// ========================================================================================

// RegisterRequest là DTO nhận từ HTTP request body khi đăng ký.
// Binding tags: Gin dùng go-playground/validator để validate.
// "required" + "email" → kiểm tra format email theo RFC 5322.
// "min=8" → enforce password complexity ở tầng đầu vào.
type RegisterRequest struct {
	Email           string `json:"email"           binding:"required,email"`
	FullName        string `json:"fullName"        binding:"required,min=2,max=100"`
	Password        string `json:"password"        binding:"required,min=8"`
	ConfirmPassword string `json:"confirmPassword" binding:"required,eqfield=Password"`
	// eqfield=Password: validator built-in để so sánh hai field.
	// Nếu không khớp → 400 Bad Request ngay tại controller, không xuống biz layer.
	// Đây là "Fail Fast" ở tầng đầu vào — bảo vệ biz khỏi invalid state.
}

// LoginRequest là DTO nhận từ HTTP request body khi đăng nhập.
type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// UserProfileResponse là DTO trả về client — chỉ có safe fields.
// KHÔNG có: password, totpSecret, googleId — ngay cả khi là hashed.
// JSON tags: camelCase theo convention của project này.
type UserProfileResponse struct {
	Id        int64   `json:"id"`
	Email     string  `json:"email"`
	FullName  *string `json:"fullName"`
	Avatar    *string `json:"avatar"`
	CreatedAt *int64  `json:"createdAt"` // Unix timestamp (số nguyên) — dễ xử lý hơn ở frontend so với ISO string
}

// AuthResponse là DTO trả về sau khi login/register thành công.
// DESIGN: Trả về token ngay sau register (auto-login) thay vì bắt user login lại.
// Trade-off: UX tốt hơn, nhưng nếu email verification là required → cần rethink flow này.
type AuthResponse struct {
	AccessToken  string              `json:"accessToken"`
	RefreshToken string              `json:"refreshToken"`
	ExpiresIn    int64               `json:"expiresIn"` // Unix timestamp — client biết khi nào cần refresh
	User         UserProfileResponse `json:"user"`
}

// RegisterResponse — sau register chỉ trả profile, chưa có token (nếu cần email verify).
// Hiện tại để đơn giản, dùng AuthResponse cho cả register lẫn login.
// TODO: Tách thành hai response khác nhau khi implement email verification.
type RegisterResponse = AuthResponse
