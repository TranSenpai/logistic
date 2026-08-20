# Luồng xác thực và phân quyền

Từ lúc người dùng đăng nhập tới lúc một request chạm được endpoint quản trị.
Đi qua **auth_service** và lớp middleware của **gateway_service**.

![Sơ đồ luồng xác thực](../diagrams/svg/authentication-flow.svg)

---

## Hai đường vào

### Đường 1 — Email + mật khẩu

```
POST /api/v1/auth/login
{ "email": "...", "password": "..." }
```

auth_service kiểm mật khẩu, ký JWT, trả về:

```json
{ "data": { "access_token": "...", "refresh_token": "...", "expires_in": 1735... } }
```

Sai mật khẩu → gRPC `Unauthenticated` → gateway dịch thành **HTTP 401**
`UNAUTHENTICATED`. (Bản cũ trả 500 cho mọi lỗi, kể cả sai mật khẩu.)

### Đường 2 — Google OAuth2

```
GET /api/v1/auth/google/login      → redirect 307 sang Google
GET /api/v1/auth/google/callback   → đổi code lấy token, set cookie
```

Cơ chế chống CSRF bằng `state` + cookie được mô tả riêng ở
[Đăng nhập Google và chống CSRF](../architecture/oauth-google-flow.md).

Điểm cốt lõi: chuỗi `state` lộ trên URL cũng vô hại, vì nó **bắt buộc phải đi kèm**
chiếc cookie `HttpOnly` chỉ tồn tại trên máy của đúng người khởi tạo yêu cầu.

---

## Gọi API sau khi có token

```
GET /api/v1/users/{user_id}
Authorization: Bearer <access_token>
```

hoặc qua cookie `access_token` (luồng OAuth đã set sẵn).

`GET /api/v1/auth/me` gọi `VerifyToken` để lấy profile hiện tại.

---

## Phân quyền ở gateway

### `IdentityContext` — chỉ chuyển thông tin, không cấp quyền

```go
X-User-Id   → ctx["ctx_user_id"]
X-User-Role → ctx["ctx_user_role"]
```

Middleware này **không tự xác thực**. Nó tin rằng lớp xác thực phía trước đã kiểm
tra. Đây là ranh giới cần nhớ rõ khi triển khai thật: nếu để header đi thẳng từ
Internet vào gateway mà không có lớp kiểm JWT, bất kỳ ai cũng tự xưng admin được.

### `RequireRole("admin")` — gắn ở CẤP GROUP

```go
admin := api.Group("/admin", middleware.RequireRole("admin"))
{
    adminUsers := admin.Group("/users")     // tự động được bảo vệ
    adminVehicles := admin.Group("/vehicles")
    // ...
}
```

Đây là điểm thiết kế quan trọng nhất của cây route: gắn quyền ở group thay vì
từng handler nghĩa là **không thể quên bảo vệ một endpoint admin mới** — chỉ cần
khai báo nó trong `adminGroup` là đã được chặn.

Có test `TestAdminRoutesAreGuarded` đếm số handler trong chuỗi để xác nhận: route
client có 7 handler, route admin có 8 (thêm `RequireRole`).

| Trường hợp | Phản hồi |
|---|---|
| Không có `X-User-Role` | 401 `UNAUTHENTICATED` |
| Role không phải `admin` | 403 `PERMISSION_DENIED` |
| Role là `admin` | Đi tiếp |

---

## Quyền ở tầng nghiệp vụ

Phân quyền ở gateway là lớp thô. Các service còn kiểm quyền sở hữu ở tầng biz:

| Kiểm tra | Ví dụ |
|---|---|
| Địa chỉ có thuộc về user này không | `ADDRESS_NOT_OWNED` |
| Thiết bị có thuộc về user này không | `DEVICE_NOT_OWNED` |
| Xe có thuộc về tài xế này không | `VEHICLE_NOT_OWNED` |
| Thông báo có thuộc về user này không | `NOTIFICATION_NOT_OWNED` |

Đoán đúng id vẫn không đọc/sửa được dữ liệu của người khác.

Các API này nhận `user_id` rỗng nghĩa là "gọi từ nội bộ hoặc admin" và bỏ qua
kiểm tra sở hữu — nên **đừng để client tự truyền `user_id` rỗng** đi qua được lớp
xác thực.

---

## Nginx: lớp phòng thủ phụ

```nginx
location /api/v1/admin/ {
    limit_req zone=api_zone burst=20 nodelay;   # siết hơn API thường
    add_header Cache-Control "no-store" always;
}
```

Đây **không** phải lớp xác thực — quyền thật sự do `RequireRole` quyết định. Nginx
chỉ siết tần suất và cấm cache.

---

## Liên quan

- [auth_service](../services/auth-service.md)
- [gateway_service](../services/gateway-service.md)
- [Đăng nhập Google và chống CSRF](../architecture/oauth-google-flow.md)
