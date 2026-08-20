package apperr

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"google.golang.org/grpc/codes"
)

// TestKindMapping khoá lại bảng dịch Kind -> gRPC code -> HTTP status.
//
// Đây là hợp đồng với client: đổi một dòng ở đây là đổi mã lỗi mà app mobile
// đang xử lý, nên phải cố ý chứ không được vô tình.
func TestKindMapping(t *testing.T) {
	cases := []struct {
		kind     Kind
		wantGRPC codes.Code
		wantHTTP int
	}{
		{KindInvalidArgument, codes.InvalidArgument, http.StatusBadRequest},
		{KindUnauthenticated, codes.Unauthenticated, http.StatusUnauthorized},
		{KindPermissionDenied, codes.PermissionDenied, http.StatusForbidden},
		{KindNotFound, codes.NotFound, http.StatusNotFound},
		{KindAlreadyExists, codes.AlreadyExists, http.StatusConflict},
		{KindConflict, codes.Aborted, http.StatusConflict},
		{KindFailedPrecond, codes.FailedPrecondition, http.StatusUnprocessableEntity},
		{KindResourceExceeded, codes.ResourceExhausted, http.StatusTooManyRequests},
		{KindUnavailable, codes.Unavailable, http.StatusServiceUnavailable},
		{KindTimeout, codes.DeadlineExceeded, http.StatusGatewayTimeout},
		{KindInternal, codes.Internal, http.StatusInternalServerError},
	}

	for _, tc := range cases {
		e := New(tc.kind, "TEST_CODE", "thông báo test")
		if got := e.GRPCCode(); got != tc.wantGRPC {
			t.Errorf("%s: gRPC code = %v, mong đợi %v", tc.kind, got, tc.wantGRPC)
		}
		if got := e.HTTPStatus(); got != tc.wantHTTP {
			t.Errorf("%s: HTTP status = %d, mong đợi %d", tc.kind, got, tc.wantHTTP)
		}
		// Vòng tròn phải khép kín: Kind -> gRPC -> HTTP cho ra cùng kết quả với
		// Kind -> HTTP. Nếu lệch, cùng một lỗi sẽ ra hai status khác nhau tuỳ
		// việc nó đi thẳng hay đi qua dây gRPC.
		if got := HTTPStatusFromGRPC(tc.wantGRPC); got != tc.wantHTTP {
			t.Errorf("%s: HTTPStatusFromGRPC(%v) = %d, mong đợi %d", tc.kind, tc.wantGRPC, got, tc.wantHTTP)
		}
	}
}

// TestFromUnwrapsThroughFmtErrorf: tầng repo hay bọc lỗi bằng fmt.Errorf("%w: ...").
// From phải xuyên qua được, nếu không thì mọi lỗi đã bọc đều rơi về Internal.
func TestFromUnwrapsThroughFmtErrorf(t *testing.T) {
	base := NotFound("USER_NOT_FOUND", "không tìm thấy người dùng")
	wrapped := fmt.Errorf("truy vấn thất bại: %w", base)
	doubleWrapped := fmt.Errorf("tầng biz: %w", wrapped)

	got, ok := From(doubleWrapped)
	if !ok {
		t.Fatal("From không bóc được lỗi qua hai lớp bọc")
	}
	if got.Code != "USER_NOT_FOUND" {
		t.Errorf("Code = %q, mong đợi USER_NOT_FOUND", got.Code)
	}
	if got.Kind != KindNotFound {
		t.Errorf("Kind = %q, mong đợi %q", got.Kind, KindNotFound)
	}
}

func TestFromRejectsForeignError(t *testing.T) {
	if _, ok := From(errors.New("lỗi lạ từ thư viện bên thứ ba")); ok {
		t.Error("From phải trả false cho lỗi không phải *apperr.Error")
	}
}

// TestWithDetailDoesNotMutateOriginal: các sentinel là biến package-level DÙNG
// CHUNG. Nếu WithDetail sửa tại chỗ, một request thêm detail sẽ làm bẩn sentinel
// cho MỌI request sau đó — kiểu lỗi rất khó lần ra.
func TestWithDetailDoesNotMutateOriginal(t *testing.T) {
	base := InvalidArgument("INVALID_ROLE", "role không hợp lệ")

	first := base.WithDetail("role", "hacker")
	second := base.WithDetail("role", "ghost")

	if len(base.Details) != 0 {
		t.Errorf("sentinel gốc bị sửa: %v", base.Details)
	}
	if first.Details["role"] != "hacker" {
		t.Errorf("bản sao thứ nhất = %v", first.Details)
	}
	if second.Details["role"] != "ghost" {
		t.Errorf("bản sao thứ hai = %v", second.Details)
	}
}

func TestWithMessageKeepsCodeAndKind(t *testing.T) {
	base := NotFound("VEHICLE_NOT_FOUND", "không tìm thấy phương tiện")
	custom := base.WithMessage("không tìm thấy xe %s", "51C-12345")

	if custom.Code != base.Code || custom.Kind != base.Kind {
		t.Errorf("WithMessage làm mất Code/Kind: %s/%s", custom.Code, custom.Kind)
	}
	if custom.Message == base.Message {
		t.Error("WithMessage không đổi được câu chữ")
	}
	if base.Message != "không tìm thấy phương tiện" {
		t.Errorf("sentinel gốc bị sửa: %q", base.Message)
	}
}

// TestCauseIsPreservedButNotInMessage: lỗi gốc phải giữ được để log, nhưng
// Message trả cho client thì không được lộ chi tiết kỹ thuật.
func TestCauseIsPreserved(t *testing.T) {
	dbErr := errors.New(`pq: duplicate key value violates unique constraint "users_phone_key"`)
	e := AlreadyExists("PHONE_ALREADY_USED", "số điện thoại đã được đăng ký").WithCause(dbErr)

	if !errors.Is(e, dbErr) {
		t.Error("errors.Is không tìm thấy lỗi gốc")
	}
	if e.Message != "số điện thoại đã được đăng ký" {
		t.Errorf("Message bị lẫn chi tiết kỹ thuật: %q", e.Message)
	}
}
