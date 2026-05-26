package entity

import "time"

// ========================================================================================
// DESIGN DECISION — Tại sao cần entity.User riêng thay vì dùng thẳng ent.Users?
//
// Ent schema (ent/schema/users.go) là INFRASTRUCTURE CONCERN — nó map thẳng vào cơ sở
// dữ liệu. Nếu bạn dùng ent.Users trực tiếp trong biz/controller, bạn đã vi phạm
// nguyên tắc Dependency Inversion (DIP trong SOLID):
//
//   Controller → Biz → [Repository Interface] ← [Ent Implementation]
//                          ↑
//                   Cái này phải là entity thuần túy (pure domain object)
//
// Lợi ích: Nếu mai mốt đổi từ Ent sang GORM, hay đổi DB từ PostgreSQL sang MongoDB,
// chỉ cần viết lại repo implementation. Controller và biz layer KHÔNG CẦN THAY ĐỔI.
// ========================================================================================

// UserProfile là entity đọc (Read model) - chứa data an toàn để trả về client.
// Lưu ý: KHÔNG có Password, TOTPSecret, GoogleID — đây là Sensitive fields theo ent schema.
// Nguyên tắc: Never serialize sensitive fields ra wire (mạng), kể cả trong internal calls.
type UserProfile struct {
	Id        int64
	Email     string
	FullName  *string
	Avatar    *string
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

// UserRegister là entity ghi (Write model / Command model) cho luồng đăng ký.
// TRADE-OFF: Tách UserRegister vs UserLogin thay vì dùng chung một struct.
// Lý do: Tránh "fat struct anti-pattern" — một struct làm nhiều việc → khó validate,
// khó mock trong test, khó bảo trì khi business rule thay đổi theo từng flow.
type UserRegister struct {
	Email    string
	FullName string
	Password string // Plain text — biz layer sẽ hash trước khi xuống repo
}

// UserLogin chỉ cần email + password để authenticate.
type UserLogin struct {
	Email    string
	Password string
}

// AuthTokenPair là kết quả trả về sau khi auth thành công.
// DESIGN CHOICE: Dùng Access Token (short-lived) + Refresh Token (long-lived).
// Pattern này là industry standard (RFC 6749 OAuth 2.0).
// Access Token: ~15 phút (stateless, verify bằng JWT signature)
// Refresh Token: ~7 ngày (stateful, lưu vào DB/Redis để có thể revoke)
//
// SCALE CONSIDERATION: Khi hệ thống đạt ~100k user:
// - Access Token verify KHÔNG cần DB hit → giảm tải đáng kể
// - Refresh Token cần 1 DB read nhưng tần suất thấp hơn nhiều (7 ngày / lần)
type AuthTokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64 // Unix timestamp khi access token hết hạn
}
