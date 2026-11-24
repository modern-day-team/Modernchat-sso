package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/modern-day-team/Modernchat-sso/server/internal/config"
)

// setupTestRouter 不传真实 DB：Setup 只构造 store，公开端点对 nil db 有默认值兜底，
// 受保护端点在到达 handler 前就被 Auth 拦截，不会查库。
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return Setup(config.Config{JWTSecret: "test-jwt-secret-0123456789abcdef-0123456789abcdef"}, nil, nil)
}

func request(r *gin.Engine, method, path string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(method, path, nil))
	return w
}

// TestPublicEndpoints 健康检查、版本、公开站点开关免鉴权。
func TestPublicEndpoints(t *testing.T) {
	r := setupTestRouter()
	for _, tc := range []struct {
		path string
		want string
	}{
		{"/healthz", `"status":"ok"`},
		{"/api/v1/version", `"version"`},
		{"/api/v1/settings", `"allowRegister"`},
	} {
		w := request(r, http.MethodGet, tc.path)
		if w.Code != http.StatusOK {
			t.Fatalf("GET %s 应 200，got %d", tc.path, w.Code)
		}
		if !strings.Contains(w.Body.String(), tc.want) {
			t.Fatalf("GET %s 响应缺 %s: %s", tc.path, tc.want, w.Body.String())
		}
	}
}

// TestProtectedEndpointsRequireAuth 认证端点无凭证必须 401（不能返回 404 或 500）。
func TestProtectedEndpointsRequireAuth(t *testing.T) {
	r := setupTestRouter()
	for _, tc := range []struct{ method, path string }{
		{http.MethodGet, "/api/v1/users/me"},
		{http.MethodPatch, "/api/v1/users/me/password"},
		{http.MethodDelete, "/api/v1/sessions/current"},
		{http.MethodGet, "/api/v1/admin/settings"},
		{http.MethodPut, "/api/v1/admin/settings"},
		{http.MethodGet, "/api/v1/admin/users"},
		{http.MethodPost, "/api/v1/admin/users/1/ban"},
		{http.MethodGet, "/api/v1/admin/update"},
	} {
		w := request(r, tc.method, tc.path)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("%s %s 应 401，got %d body=%s", tc.method, tc.path, w.Code, w.Body.String())
		}
	}
}

// TestNotFoundRoute 未注册路径必须 404（避免误把鉴权失败当路由缺失）。
func TestNotFoundRoute(t *testing.T) {
	w := request(setupTestRouter(), http.MethodGet, "/api/v1/nonexistent")
	if w.Code != http.StatusNotFound {
		t.Fatalf("未知路径应 404，got %d", w.Code)
	}
}
