package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/modern-day-team/Modernchat-sso/server/internal/apperr"
	"github.com/modern-day-team/Modernchat-sso/server/internal/model"
	"github.com/modern-day-team/Modernchat-sso/server/internal/token"
)

// fakeAuthenticator 用可控返回驱动 Auth 的各分支，避免依赖真实 service/DB。
type fakeAuthenticator struct {
	user *model.User
	err  error
	got  string
}

func (f *fakeAuthenticator) Authenticate(raw string) (*model.User, error) {
	f.got = raw
	return f.user, f.err
}

func newAuthRouter(f *fakeAuthenticator) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/protected", Auth(f), func(c *gin.Context) {
		u := UserFrom(c)
		if u == nil {
			c.JSON(http.StatusOK, gin.H{"uid": 0})
			return
		}
		c.JSON(http.StatusOK, gin.H{"uid": u.ID})
	})
	return r
}

// TestAuthRequiresBearer 无凭证必须 401（token 不能靠猜路径绕过）。
func TestAuthRequiresBearer(t *testing.T) {
	f := &fakeAuthenticator{}
	w := httptest.NewRecorder()
	newAuthRouter(f).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/protected", nil))
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("无凭证应 401，got %d", w.Code)
	}
	if f.got != "" {
		t.Fatalf("无凭证不应调用 Authenticate，got %q", f.got)
	}
}

// TestAuthAcceptsBearerToken 合法 Bearer token 必须解析出用户。
func TestAuthAcceptsBearerToken(t *testing.T) {
	f := &fakeAuthenticator{user: &model.User{ID: 42}}
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	newAuthRouter(f).ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("合法 token 应 200，got %d", w.Code)
	}
	if f.got != "valid-token" {
		t.Fatalf("Authenticate 应收到 token，got %q", f.got)
	}
	if !strings.Contains(w.Body.String(), `"uid":`+strconv.FormatUint(42, 10)) {
		t.Fatalf("响应应含用户 ID，got %s", w.Body.String())
	}
}

// TestAuthAcceptsCookieFallback <img>/WS 场景靠 cookie 兜底，必须可用。
func TestAuthAcceptsCookieFallback(t *testing.T) {
	f := &fakeAuthenticator{user: &model.User{ID: 7}}
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: token.AuthCookieName, Value: "cookie-token"})
	newAuthRouter(f).ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("cookie token 应 200，got %d", w.Code)
	}
	if f.got != "cookie-token" {
		t.Fatalf("应使用 cookie token，got %q", f.got)
	}
}

// TestAuthRejectsBannedUser 封禁用户必须 403。
func TestAuthRejectsBannedUser(t *testing.T) {
	f := &fakeAuthenticator{err: apperr.ErrBanned}
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer banned-token")
	newAuthRouter(f).ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("封禁用户应 403，got %d", w.Code)
	}
}

// TestAuthRejectsInvalidToken 无效 token 必须 401。
func TestAuthRejectsInvalidToken(t *testing.T) {
	f := &fakeAuthenticator{err: apperr.ErrInvalidToken}
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer bad-token")
	newAuthRouter(f).ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("无效 token 应 401，got %d", w.Code)
	}
}

// TestAuthUnknownErrorIs500 未知错误（DB 故障）必须 500，不能误判为 401。
func TestAuthUnknownErrorIs500(t *testing.T) {
	f := &fakeAuthenticator{err: errors.New("db down")}
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer some-token")
	newAuthRouter(f).ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("未知错误应 500，got %d", w.Code)
	}
}

// Admin 角色校验：普通用户必须 403，站长放行。
func TestAdminRequiresAdminRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name string
		role string
		want int
	}{
		{name: "admin", role: "admin", want: http.StatusOK},
		{name: "user", role: "user", want: http.StatusForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := &fakeAuthenticator{user: &model.User{ID: 1, Role: tc.role}}
			r := gin.New()
			r.GET("/admin", Auth(f), Admin(), func(c *gin.Context) { c.Status(http.StatusOK) })
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/admin", nil)
			req.Header.Set("Authorization", "Bearer token")
			r.ServeHTTP(w, req)
			if w.Code != tc.want {
				t.Fatalf("role=%s 期望 %d，got %d", tc.role, tc.want, w.Code)
			}
		})
	}
}
