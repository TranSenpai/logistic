# gateway_service

Cửa duy nhất mà thế giới bên ngoài chạm được. Dịch HTTP ↔ gRPC, chặn quyền quản
trị, chuẩn hoá mọi phản hồi. **Không có database, không có luật nghiệp vụ.**

| | |
|---|---|
| Cổng | 8080 |
| Giao thức vào | HTTP (qua Nginx) |
| Giao thức ra | gRPC tới 6 service nội bộ |
| Database | không |
| Số endpoint | 67 (49 client + 18 admin) |

![Sơ đồ gateway_service](../diagrams/svg/gateway-service.svg)

## Trách nhiệm

1. **Định tuyến** — cây route `/api/v1/...` (client) và `/api/v1/admin/...` (quản trị).
2. **Phân quyền** — `RequireRole("admin")` gắn ở **cấp group**, không phải từng handler.
3. **Chuẩn hoá lỗi** — đọc gRPC status + `ErrorInfo` rồi dịch sang HTTP status + mã lỗi ổn định.
4. **Truy vết** — sinh `X-Request-ID` và trả lại trong mọi phản hồi.

## Chuỗi middleware

Thứ tự đăng ký **có ý nghĩa**:

| Thứ tự | Middleware | Vì sao ở vị trí này |
|---|---|---|
| 1 | `RequestID` | Phải chạy đầu để mọi log phía sau có mã truy vết |
| 2 | `Recovery` | Bọc ngoài các middleware còn lại để bắt được cả panic của chúng |
| 3 | `AccessLog` | Ghi method, path, status, thời gian |
| 4 | `IdentityContext` | Đọc `X-User-Id` / `X-User-Role` vào context |
| 5 | `ErrorGuard` | Lưới an toàn: render lỗi mà handler quên render |

Dùng `gin.New()` chứ **không** `gin.Default()`: Logger và Recovery mặc định của
gin ghi log dạng text và trả body 500 không theo khung lỗi chung của hệ thống.

## Khung phản hồi

Thành công:

```json
{ "data": { }, "message": "…", "request_id": "018f…" }
```

Lỗi:

```json
{ "error": { "code": "USER_NOT_FOUND", "message": "…", "details": { } },
  "request_id": "018f…" }
```

Client nên switch theo `error.code` — mã này ổn định theo thời gian, còn
`message` là câu chữ có thể đổi.

## Quyết định thiết kế đáng chú ý

**Vì sao controller không có một dòng `status.Errorf` nào?**
Service nội bộ đã gắn `ErrorInfo` vào gRPC status (xem `pkg/middleware`).
`response.Error(ctx, err)` bóc ra và tự chọn HTTP status. Trước đây mỗi controller
tự viết `ctx.JSON(500, …)`, hệ quả là **mọi** lỗi — kể cả "không tìm thấy user" —
đều ra HTTP 500.

**Vì sao ID trong proto của user/vehicle/notification là `string` chứ không phải `bytes`?**
Bản cũ dùng `bytes` (16 byte thô của UUID), và gateway viết `[]byte(id)` — biến
chuỗi 36 ký tự thành 36 byte, tức là gửi xuống một UUID sai hoàn toàn. Chuyển sang
`string` dạng RFC-4122 thì đọc được bằng mắt, đi thẳng vào URL, và việc parse chỉ
xảy ra đúng một chỗ ở tầng mapper.

`matching_service` vẫn giữ `bytes`, nên gateway phải parse tường minh qua
`uuidBytes()` — xem `matching_controller.go`.

## Cấu hình

```
GATEWAY_PORT=8080
GATEWAY_AUTH_GRPC_ADDR=auth-service:9001
GATEWAY_MEDIA_GRPC_ADDR=media-service:9002
GATEWAY_MATCHING_GRPC_ADDR=matching-service:9003
GATEWAY_USER_GRPC_ADDR=user-service:9004
GATEWAY_VEHICLE_GRPC_ADDR=vehicle-service:9005
GATEWAY_NOTIFICATION_GRPC_ADDR=notification-service:9006
```

`grpc.NewClient` là lazy — không bắt tay ngay mà kết nối ở lần gọi đầu. Nhờ vậy
gateway khởi động được kể cả khi service phía sau chưa sẵn sàng.

## Liên quan

- [Luồng xác thực và phân quyền](../flows/authentication-flow.md)
- [Luồng lỗi xuyên tầng](../flows/error-handling-flow.md)
