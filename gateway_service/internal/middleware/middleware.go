// Package middleware chứa các lớp chặn HTTP của gateway.
//
// Thứ tự đăng ký trong gateway_route.go có ý nghĩa:
//
//	RequestID  -> sinh mã truy vết, phải chạy đầu để mọi log sau đó có mã.
//	Recovery   -> bắt panic, phải bọc ngoài các middleware còn lại.
//	ErrorGuard -> chốt chặn cuối, render lỗi mà controller quên render.
package middleware

import (
	"log"
	"runtime/debug"
	"time"

	"gateway_service/internal/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestID gắn X-Request-ID cho mỗi request.
//
// Client gửi sẵn thì dùng lại (giữ được chuỗi truy vết xuyên nhiều hệ thống),
// không thì sinh mới. Mã này được trả lại trong header VÀ trong body lỗi, nên
// khi người dùng báo lỗi chỉ cần đưa mã là tra được đúng dòng log.
func RequestID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.GetHeader(response.HeaderRequestID)
		if id == "" {
			id = uuid.Must(uuid.NewV7()).String()
		}
		ctx.Set(response.HeaderRequestID, id)
		ctx.Writer.Header().Set(response.HeaderRequestID, id)
		ctx.Next()
	}
}

// Recovery biến panic thành HTTP 500 có cấu trúc thay vì làm sập gateway.
//
// gin có Recovery riêng, nhưng nó trả về body rỗng/plain-text không khớp với
// khung lỗi của hệ thống. Client phải parse được lỗi 500 giống mọi lỗi khác.
func Recovery() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[gateway][PANIC] %s %s -> %v\n%s",
					ctx.Request.Method, ctx.Request.URL.Path, r, debug.Stack())

				if !ctx.Writer.Written() {
					ctx.AbortWithStatusJSON(500, response.ErrorBody{
						Error: response.ErrorDetail{
							Code:    "INTERNAL",
							Message: "internal gateway error",
						},
						RequestID: ctx.GetString(response.HeaderRequestID),
					})
				}
			}
		}()
		ctx.Next()
	}
}

// ErrorGuard là lưới an toàn: nếu một handler gọi ctx.Error(err) rồi return mà
// chưa ghi gì ra response, ta render lỗi đó ở đây thay vì trả 200 body rỗng.
func ErrorGuard() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()

		if len(ctx.Errors) == 0 || ctx.Writer.Written() {
			return
		}
		response.Error(ctx, ctx.Errors.Last().Err)
	}
}

// AccessLog ghi lại method, path, status và thời gian xử lý.
func AccessLog() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		ctx.Next()

		log.Printf("[gateway] %s %s -> %d (%s) req_id=%s",
			ctx.Request.Method,
			ctx.Request.URL.Path,
			ctx.Writer.Status(),
			time.Since(start),
			ctx.GetString(response.HeaderRequestID),
		)
	}
}

// ---------------------------------------------------------------------------
// PHÂN QUYỀN
// ---------------------------------------------------------------------------

// HeaderUserID / HeaderUserRole là các header do auth_service (hoặc lớp xác thực
// phía trước) gắn vào sau khi giải mã JWT.
const (
	HeaderUserID   = "X-User-Id"
	HeaderUserRole = "X-User-Role"

	CtxUserID   = "ctx_user_id"
	CtxUserRole = "ctx_user_role"
)

// IdentityContext đọc danh tính từ header vào gin context để controller dùng.
//
// KHÔNG tự xác thực ở đây: gateway tin phần đầu vào đã được lớp auth kiểm tra.
// Middleware này chỉ chuyển thông tin xuống, không cấp quyền cho ai.
func IdentityContext() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if uid := ctx.GetHeader(HeaderUserID); uid != "" {
			ctx.Set(CtxUserID, uid)
		}
		if role := ctx.GetHeader(HeaderUserRole); role != "" {
			ctx.Set(CtxUserRole, role)
		}
		ctx.Next()
	}
}

// RequireRole chặn nhóm route quản trị.
//
// Đây là lý do nhóm /api/v1/admin được tách hẳn khỏi nhóm client: chỉ cần gắn
// middleware này MỘT lần ở cấp group là toàn bộ endpoint admin được bảo vệ,
// thay vì phải nhớ kiểm tra quyền trong từng handler.
func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}

	return func(ctx *gin.Context) {
		role := ctx.GetString(CtxUserRole)
		if role == "" {
			response.Unauthorized(ctx, "thiếu thông tin xác thực")
			return
		}
		if _, ok := allowed[role]; !ok {
			response.Forbidden(ctx, "tài khoản không có quyền truy cập khu vực quản trị")
			return
		}
		ctx.Next()
	}
}

// CurrentUserID trả về id người dùng đang gọi (rỗng nếu chưa xác thực).
// Controller dùng nó để mặc định user_id khi client không truyền tường minh.
func CurrentUserID(ctx *gin.Context) string {
	return ctx.GetString(CtxUserID)
}
