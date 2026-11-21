// Package handler HTTP 处理层：绑定请求 → 调用账号业务（窄接口）→ 映射状态码与响应。
// 业务规则一律在 service；本包不 import store，不写账号逻辑。
package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/modern-day-team/Modernchat-sso/server/internal/apperr"
	"github.com/modern-day-team/Modernchat-sso/server/internal/middleware"
	"github.com/modern-day-team/Modernchat-sso/server/internal/model"
	"github.com/modern-day-team/Modernchat-sso/server/internal/token"
)

// AccountService 消费方定义的窄接口（由 service.AccountService 实现）。
type AccountService interface {
	Register(username, password, nickname string, email *string) (*model.User, error)
	Login(username, password, ip string) (tokenStr string, user *model.User, err error)
	Logout(userID uint) error
}

// AuthHandler 认证相关 HTTP 端点（RESTful：POST /sessions 即登录）。
type AuthHandler struct {
	accounts     AccountService
	cookieSecure bool
}

// NewAuthHandler 构造。
func NewAuthHandler(accounts AccountService, cookieSecure bool) *AuthHandler {
	return &AuthHandler{accounts: accounts, cookieSecure: cookieSecure}
}

type registerReq struct {
	Username string  `json:"username" binding:"required,max=20"`
	Password string  `json:"password" binding:"required,min=6,max=72"`
	Nickname string  `json:"nickname" binding:"max=20"`
	Email    *string `json:"email" binding:"omitempty,email,max=254"`
}

// Register 注册。业务规则（首个 admin、唯一性、默认头像色）在 service。
func (h *AuthHandler) Register(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "参数不合法：用户名 3-20 位字母/数字/_/-，密码 6-72 位")
		return
	}
	u, err := h.accounts.Register(req.Username, req.Password, req.Nickname, req.Email)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, u)
}

type loginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login 登录：签发 token 并种 mc_auth cookie。
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "参数不合法")
		return
	}
	tok, u, err := h.accounts.Login(req.Username, req.Password, c.ClientIP())
	if err != nil {
		mapErr(c, err)
		return
	}
	setAuthCookie(c, h.cookieSecure, tok)
	c.JSON(http.StatusOK, gin.H{"token": tok, "user": u})
}

// Logout 登出：吊销该用户全部已签发 token（service 内 bump token_version），并清 cookie。
func (h *AuthHandler) Logout(c *gin.Context) {
	u := middleware.UserFrom(c)
	if u == nil {
		fail(c, http.StatusUnauthorized, "未登录")
		return
	}
	if err := h.accounts.Logout(u.ID); err != nil {
		mapErr(c, err)
		return
	}
	clearAuthCookie(c, h.cookieSecure)
	c.Status(http.StatusNoContent)
}

// Me 当前用户。
func (h *AuthHandler) Me(c *gin.Context) {
	c.JSON(http.StatusOK, middleware.UserFrom(c))
}

// mapErr 领域错误 → HTTP 状态码：本层唯一的错误映射点，依赖 apperr 哨兵而非 service 实现。
func mapErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, apperr.ErrInvalidInput):
		fail(c, http.StatusBadRequest, "参数不合法：用户名 3-20 位字母/数字/_/-，密码 6-72 位")
	case errors.Is(err, apperr.ErrInvalidEmail):
		fail(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, apperr.ErrUsernameTaken), errors.Is(err, apperr.ErrEmailTaken):
		fail(c, http.StatusConflict, err.Error())
	case errors.Is(err, apperr.ErrInvalidCredentials), errors.Is(err, apperr.ErrInvalidToken):
		fail(c, http.StatusUnauthorized, err.Error())
	case errors.Is(err, apperr.ErrBanned):
		fail(c, http.StatusForbidden, err.Error())
	default:
		slog.Error("内部错误", "err", err, "path", c.FullPath())
		// 500 文案与主仓库规格 §5 一致；真实原因只进日志
		fail(c, http.StatusInternalServerError, "服务器开小差了，请稍后再试")
	}
}

func setAuthCookie(c *gin.Context, secure bool, tok string) {
	c.SetSameSite(http.SameSiteLaxMode)
	// TODO：SSO 跨子域部署时需加 Domain 并改 SameSite=None + COOKIE_SECURE=true
	c.SetCookie(token.AuthCookieName, tok, int(token.TTL.Seconds()), "/", "", secure, true)
}

func clearAuthCookie(c *gin.Context, secure bool) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(token.AuthCookieName, "", -1, "/", "", secure, true)
}

func fail(c *gin.Context, code int, msg string) {
	c.JSON(code, gin.H{"error": msg})
}
