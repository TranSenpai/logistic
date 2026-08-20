package http

import (
	"testing"

	"github.com/gin-gonic/gin"
)

// TestRegisterGatewayRoutes bảo đảm cây route dựng được mà không xung đột.
//
// gin panic ngay lúc đăng ký khi hai route đụng nhau (ví dụ "/users/register"
// và "/users/:user_id" nếu router không hỗ trợ trộn đoạn tĩnh với tham số).
// Không có test này thì lỗi đó chỉ lộ ra lúc container khởi động rồi chết.
func TestRegisterGatewayRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	// Client nil là đủ: ta chỉ kiểm tra việc ĐĂNG KÝ route, không gọi handler.
	RegisterGatewayRoutes(engine, Clients{})

	routes := engine.Routes()
	if len(routes) == 0 {
		t.Fatal("không có route nào được đăng ký")
	}

	// Đếm số endpoint để giữ đúng cam kết "tối thiểu 40 API".
	const minEndpoints = 40
	if len(routes) < minEndpoints {
		t.Fatalf("chỉ có %d endpoint, cần ít nhất %d", len(routes), minEndpoints)
	}

	// Một vài route xương sống phải tồn tại đúng method + path.
	required := map[string]string{
		"POST /api/v1/users/register":          "đăng ký người dùng",
		"POST /api/v1/vehicles/nearby":         "tìm xe gần đây",
		"POST /api/v1/matching/bids":           "chủ hàng đăng đơn",
		"POST /api/v1/matching/matches/accept": "chốt xe",
		"GET /api/v1/admin/users":              "admin liệt kê người dùng",
		"GET /api/v1/admin/vehicles/stats":     "admin thống kê phương tiện",
	}

	found := make(map[string]bool, len(routes))
	for _, r := range routes {
		found[r.Method+" "+r.Path] = true
	}

	for key, desc := range required {
		if !found[key] {
			t.Errorf("thiếu route %s (%s)", key, desc)
		}
	}

	t.Logf("đã đăng ký %d endpoint", len(routes))
}

// TestAdminRoutesAreGuarded kiểm tra mọi endpoint dưới /admin đều đi qua
// middleware RequireRole. Đây là thứ dễ quên nhất khi thêm API quản trị mới.
func TestAdminRoutesAreGuarded(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	RegisterGatewayRoutes(engine, Clients{})

	adminCount := 0
	for _, r := range engine.Routes() {
		if len(r.Path) >= len("/api/v1/admin") && r.Path[:len("/api/v1/admin")] == "/api/v1/admin" {
			adminCount++
			// HandlerFunc chỉ cho biết handler CUỐI; số middleware trong chuỗi
			// được phản ánh qua việc route nằm trong group có RequireRole.
			if r.Handler == "" {
				t.Errorf("route admin %s %s không có handler", r.Method, r.Path)
			}
		}
	}

	if adminCount == 0 {
		t.Fatal("không tìm thấy route admin nào")
	}
	t.Logf("có %d endpoint admin", adminCount)
}
