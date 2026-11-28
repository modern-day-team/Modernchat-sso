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
	"github.com/modern-day-team/Modernchat-sso/server/internal/service"
	"github.com/modern-day-team/Modernchat-sso/server/internal/token"
)

// AccountService 消费方定义的窄接口（由 service.AccountService 实现）。
type AccountService interface {
	Register(username, password, nickname, email, emailCode string) (*model.User, error)
	Login(username, password, ip string) (tokenStr string, user *model.User, err error)
	Logout(userID uint) error
	SendEmailCode(email, purpose string) (cooldownSeconds int, err error)
	Forgot(email string)
	Reset(token, newPassword string) error
	ChangePassword(userID uint, oldPassword, newPassword string) error
}

// AuthHandler 认证相关 HTTP 端点（RESTful：POST /sessions 即登录）。
type AuthHandler struct {
	accounts     AccountService
	cookieSecure bool
	cookieDomain string // 跨子域共享时为父域（如 .example.com）；留空=当前域
}

// NewAuthHandler 构造。cookieDomain 非空时走跨子域模式（SameSite=None + Domain）。
func NewAuthHandler(accounts AccountService, cookieSecure bool, cookieDomain string) *AuthHandler {
	return &AuthHandler{accounts: accounts, cookieSecure: cookieSecure, cookieDomain: cookieDomain}
}

type registerReq struct {
	Username  string `json:"username" binding:"required,max=20"`
	Password  string `json:"password" binding:"required,min=6,max=72"`
	Nickname  string `json:"nickname" binding:"max=20"`
	Email     string `json:"email" binding:"omitempty,email,max=254"`
	EmailCode string `json:"emailCode"`
}

// Register 注册。业务规则（首个 admin、唯一性、默认头像色）在 service。
func (h *AuthHandler) Register(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "参数不合法：用户名 3-20 位字母/数字/_/-，密码 6-72 位")
		return
	}
	u, err := h.accounts.Register(req.Username, req.Password, req.Nickname, req.Email, req.EmailCode)
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
	setAuthCookie(c, h.cookieSecure, h.cookieDomain, tok)
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
	clearAuthCookie(c, h.cookieSecure, h.cookieDomain)
	c.Status(http.StatusNoContent)
}

// Me 当前用户。
func (h *AuthHandler) Me(c *gin.Context) {
	c.JSON(http.StatusOK, middleware.UserFrom(c))
}

type emailCodeReq struct {
	Email   string `json:"email" binding:"required"`
	Purpose string `json:"purpose"`
}

// SendEmailCode POST /api/v1/email-codes：生成并发送邮箱验证码（skipAuth 公开）。
// 响应 {ok, cooldownSeconds}：冷却中返回剩余+1、成功返回 60（防枚举，主仓库同款）。
func (h *AuthHandler) SendEmailCode(c *gin.Context) {
	var req emailCodeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "请填写邮箱")
		return
	}
	purpose := req.Purpose
	switch purpose {
	case service.EmailPurposeRegister, service.EmailPurposeChange:
	default:
		purpose = service.EmailPurposeRegister
	}
	cooldown, err := h.accounts.SendEmailCode(req.Email, purpose)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "cooldownSeconds": cooldown})
}

type forgotReq struct {
	Email string `json:"email" binding:"required"`
}

// Forgot POST /api/v1/password-resets：恒 200 {ok:true}，不暴露邮箱存在性
// （邮件发送失败也静默——对齐主仓库 Forgot 的防枚举语义）。
func (h *AuthHandler) Forgot(c *gin.Context) {
	var req forgotReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "请填写邮箱")
		return
	}
	h.accounts.Forgot(req.Email)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

type resetReq struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=6,max=72"`
}

// Reset POST /api/v1/password-resets/:token：重置密码（令牌单次消费 + tv bump 全端失效）。
func (h *AuthHandler) Reset(c *gin.Context) {
	var req resetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "参数不合法：密码 6-72 位")
		return
	}
	if err := h.accounts.Reset(req.Token, req.NewPassword); err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

type changePasswordReq struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=6,max=72"`
}

// ChangePassword PATCH /api/v1/users/me/password：自身改密（SSO 新增，主仓库本期无独立入口）。
// 改密后 token_version+1 全端失效，客户端应清本地 token 并引导重新登录。
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	u := middleware.UserFrom(c)
	if u == nil {
		fail(c, http.StatusUnauthorized, "未登录")
		return
	}
	var req changePasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "参数不合法：密码 6-72 位")
		return
	}
	if err := h.accounts.ChangePassword(u.ID, req.OldPassword, req.NewPassword); err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// mapErr 领域错误 → HTTP 状态码：本层唯一的错误映射点，依赖 apperr 哨兵而非 service 实现。
func mapErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, apperr.ErrInvalidInput):
		fail(c, http.StatusBadRequest, "参数不合法：用户名 3-20 位字母/数字/_/-，密码 6-72 位")
	case errors.Is(err, apperr.ErrInvalidEmail), errors.Is(err, apperr.ErrEmailRequired), errors.Is(err, apperr.ErrEmailCodeRequired):
		fail(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, apperr.ErrUsernameTaken), errors.Is(err, apperr.ErrEmailTaken):
		fail(c, http.StatusConflict, err.Error())
	case errors.Is(err, apperr.ErrEmailCodeInvalid):
		// 422：语义上有害输入（验证码错误），对齐主仓库规格
		fail(c, http.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, apperr.ErrInvalidCredentials), errors.Is(err, apperr.ErrInvalidToken):
		fail(c, http.StatusUnauthorized, err.Error())
	case errors.Is(err, apperr.ErrBanned):
		fail(c, http.StatusForbidden, err.Error())
	case errors.Is(err, apperr.ErrRegisterClosed):
		fail(c, http.StatusForbidden, err.Error())
	case errors.Is(err, apperr.ErrAccountLocked):
		fail(c, http.StatusTooManyRequests, err.Error())
	case errors.Is(err, apperr.ErrMailNotReady), errors.Is(err, apperr.ErrMailSendFailed):
		fail(c, http.StatusServiceUnavailable, err.Error())
	case errors.Is(err, apperr.ErrUserNotFound):
		fail(c, http.StatusNotFound, err.Error())
	case errors.Is(err, apperr.ErrInvalidDays), errors.Is(err, apperr.ErrInvalidRole), errors.Is(err, apperr.ErrOldPassword), errors.Is(err, apperr.ErrResetTokenStale):
		fail(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, apperr.ErrCannotBanAdmin), errors.Is(err, apperr.ErrLastAdmin):
		fail(c, http.StatusForbidden, err.Error())
	case errors.Is(err, apperr.ErrResetTokenInvalid):
		// 422：与验证码错误同级的「有害输入」语义
		fail(c, http.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, apperr.ErrResetTokenStale), errors.Is(err, apperr.ErrOldPassword):
		fail(c, http.StatusBadRequest, err.Error())
	default:
		slog.Error("内部错误", "err", err, "path", c.FullPath())
		// 500 文案与主仓库规格 §5 一致；真实原因只进日志
		fail(c, http.StatusInternalServerError, "服务器开小差了，请稍后再试")
	}
}

func setAuthCookie(c *gin.Context, secure bool, domain, tok string) {
	if domain != "" {
		// 跨子域共享：Domain 指向父域 + SameSite=None（浏览器仅接受带 Secure 的 None，
		// config.Load 已强制 COOKIE_SECURE=true，这里 domain 非空即 secure）。
		c.SetSameSite(http.SameSiteNoneMode)
		c.SetCookie(token.AuthCookieName, tok, int(token.TTL.Seconds()), "/", domain, true, true)
		return
	}
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(token.AuthCookieName, tok, int(token.TTL.Seconds()), "/", "", secure, true)
}

func clearAuthCookie(c *gin.Context, secure bool, domain string) {
	if domain != "" {
		c.SetSameSite(http.SameSiteNoneMode)
		c.SetCookie(token.AuthCookieName, "", -1, "/", domain, true, true)
		return
	}
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(token.AuthCookieName, "", -1, "/", "", secure, true)
}

func fail(c *gin.Context, code int, msg string) {
	c.JSON(code, gin.H{"error": msg})
}
