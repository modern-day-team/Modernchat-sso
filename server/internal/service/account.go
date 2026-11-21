// Package service 业务层：账号/认证业务规则唯一归属地。
// 只依赖 store、token、apperr、mailer（本包），不感知 HTTP（gin）；错误一律返回 apperr 哨兵。
package service

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/modern-day-team/Modernchat-sso/server/internal/apperr"
	"github.com/modern-day-team/Modernchat-sso/server/internal/model"
	"github.com/modern-day-team/Modernchat-sso/server/internal/store"
	"github.com/modern-day-team/Modernchat-sso/server/internal/token"
)

// 邮件验证用途（业务概念；store 层只收字符串）。
const (
	EmailPurposeRegister = "register"
	EmailPurposeChange   = "change_email"
)

// usernamePattern 用户名规则与主仓库一致：3-20 位，仅字母/数字/下划线/连字符（规格 §4）。
var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,20}$`)

// 邮箱规则与主仓库一致（规格 §4）：^[^\s@]+@[^\s@]+\.[^\s@]+$，≤254 字符
var emailPattern = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

// 头像默认取色板（与主仓库 TDesign 色系一致）。
var avatarPalette = []string{"#0052d9", "#2ba471", "#d54941", "#ed7b2f", "#ca8d04"}

// AccountService 账号业务。
type AccountService struct {
	users         *store.UserStore
	emailCodes    *store.EmailCodeStore
	loginAttempts *store.LoginAttemptStore
	settings      *store.SettingStore
	mailer        Mailer
	secret        string // JWT 签名密钥（与主仓库共享同一值即 token 互认）
}

// NewAccountService 构造。emailCodes/mailer/loginAttempts/settings 允许 nil，对应能力降级：
// 无邮件通道时注册邮箱验证走 503 降级守卫（首账号豁免），无锁定表时不锁定。
func NewAccountService(users *store.UserStore, emailCodes *store.EmailCodeStore, mailer Mailer,
	loginAttempts *store.LoginAttemptStore, settings *store.SettingStore, secret string) *AccountService {
	return &AccountService{
		users: users, emailCodes: emailCodes, mailer: mailer,
		loginAttempts: loginAttempts, settings: settings, secret: secret,
	}
}

// ops 运行开关（settings 缺失时回退默认值）。
func (s *AccountService) ops() store.Ops {
	if s.settings == nil {
		return store.DefaultOps()
	}
	return s.settings.Ops()
}

// mailerReady 邮件服务是否就绪（对齐主仓库的单一门禁）。
func (s *AccountService) mailerReady() bool {
	return s.mailer != nil && s.mailer.Enabled()
}

// Register 注册：参数校验、站点开关、邮箱验证码、首个注册者为 admin（沿用主仓库约定）、默认头像色。
func (s *AccountService) Register(username, password, nickname string, email, emailCode string) (*model.User, error) {
	if !usernamePattern.MatchString(username) || len(password) < 6 || len(password) > 72 {
		return nil, apperr.ErrInvalidInput
	}
	ops := s.ops()
	if !ops.AllowRegister {
		return nil, apperr.ErrRegisterClosed
	}
	if ops.RequireEmailVerify && email == "" {
		return nil, apperr.ErrEmailRequired
	}
	if email != "" {
		email = normalizeEmail(email)
		if !validEmail(email) {
			return nil, apperr.ErrInvalidEmail
		}
		if _, err := s.users.FindByEmail(email); err == nil {
			return nil, apperr.ErrEmailTaken
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}
	count, err := s.users.Count()
	if err != nil {
		return nil, err
	}
	mailerReady := s.mailerReady()
	// 首账号豁免验证码（部署引导）；RequireEmailVerify 开启且非首账号时强制（主仓库同款）。
	needCode := ops.RequireEmailVerify && mailerReady && count > 0
	if ops.RequireEmailVerify && !mailerReady && count > 0 {
		// 降级守卫：不能因为邮件没配好就绕过验证策略批量注册（主仓库同款）
		return nil, apperr.ErrMailNotReady
	}
	if needCode {
		if emailCode == "" {
			return nil, apperr.ErrEmailCodeRequired
		}
		ok, err := s.emailCodes.Verify(email, EmailPurposeRegister, emailCode)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, apperr.ErrEmailCodeInvalid
		}
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	u := model.User{
		Username:      username,
		PasswordHash:  string(hash),
		Nickname:      cmp.Or(nickname, username),
		AvatarColor:   avatarPalette[int(count)%len(avatarPalette)],
		EmailVerified: needCode,
	}
	if email != "" {
		u.Email = &email
	}
	if count == 0 {
		u.Role = "admin" // 首个注册者即为管理员
	}
	if err := s.users.Create(&u); err != nil {
		// 用户名/邮箱唯一索引兜底并发场景
		return nil, err
	}
	return &u, nil
}

// SendEmailCode 生成并发送邮箱验证码；返回冷却秒数
// （冷却期内返回剩余+1，防枚举——与主仓库一致）。
func (s *AccountService) SendEmailCode(email, purpose string) (int, error) {
	if !s.mailerReady() {
		return 0, apperr.ErrMailNotReady
	}
	email = normalizeEmail(email)
	if !validEmail(email) {
		return 0, apperr.ErrInvalidEmail
	}
	code, retryAfter, err := s.emailCodes.Issue(email, purpose)
	if err != nil {
		return 0, err
	}
	if code == "" {
		return int(retryAfter.Seconds()) + 1, nil
	}
	subject, body := emailCodeMessage(purpose, code)
	if err := s.mailer.Send(context.Background(), MailMessage{To: email, Subject: subject, Text: body}); err != nil {
		// 家族规则：日志不含验证码正文
		slog.Error("验证码邮件发送失败", "err", err, "purpose", purpose)
		return 0, apperr.ErrMailSendFailed
	}
	return int(store.EmailCodeCooldown.Seconds()), nil
}

// emailCodeMessage 验证码邮件文案（对齐主仓库 mailer.go SendVerifyCode）。
func emailCodeMessage(purpose, code string) (subject, body string) {
	switch purpose {
	case EmailPurposeRegister:
		subject = "Modernchat SSO 注册验证码"
		body = fmt.Sprintf("你的注册验证码是：%s\n有效期 10 分钟，仅可使用一次，请勿转发给他人。\n若非本人操作，请忽略本邮件。", code)
	default:
		subject = "Modernchat SSO 邮箱验证码"
		body = fmt.Sprintf("你的验证码是：%s\n有效期 10 分钟，仅可使用一次，请勿转发给他人。\n若非本人操作，请忽略本邮件。", code)
	}
	return subject, body
}

// Login 登录：账号级锁定 → 校验凭证与封禁 → 签发 token。
// 家族规则：登录失败只记用户名与 IP，绝不记密码/令牌；账号不存在与密码错误同一错误。
func (s *AccountService) Login(username, password, ip string) (string, *model.User, error) {
	ops := s.ops()
	if s.loginAttempts != nil && ops.LoginMaxAttempts > 0 {
		lockedUntil, err := s.loginAttempts.IsLocked(username)
		if err != nil {
			return "", nil, err
		}
		if lockedUntil.After(time.Now()) {
			// 向上取整分钟，与主仓库「请 N 分钟后重试」一致
			mins := int(time.Until(lockedUntil).Minutes()) + 1
			return "", nil, fmt.Errorf("%w，请 %d 分钟后重试", apperr.ErrAccountLocked, mins)
		}
	}
	u, err := s.users.FindByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Warn("登录失败", "username", username, "ip", ip)
			return "", nil, apperr.ErrInvalidCredentials
		}
		return "", nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		slog.Warn("登录失败", "username", username, "ip", ip)
		if s.loginAttempts != nil && ops.LoginMaxAttempts > 0 {
			if ferr := s.loginAttempts.Fail(username, ops.LoginMaxAttempts, time.Duration(ops.LoginLockMinutes)*time.Minute); ferr != nil {
				slog.Warn("登录失败计数写入失败", "err", ferr)
			}
		}
		return "", nil, apperr.ErrInvalidCredentials
	}
	if u.IsBanned() {
		return "", nil, bannedErr(u.BanReason)
	}
	if s.loginAttempts != nil {
		if rerr := s.loginAttempts.Reset(username); rerr != nil {
			slog.Warn("登录失败计数清零失败", "err", rerr)
		}
	}
	t, err := token.Issue(u.ID, u.TokenVersion, s.secret)
	if err != nil {
		return "", nil, err
	}
	return t, u, nil
}

// Logout 登出：token_version + 1，使该用户已签发的全部 token 失效。
func (s *AccountService) Logout(userID uint) error {
	return s.users.BumpTokenVersion(userID)
}

// Authenticate 把原始凭证（JWT）解析为当前用户，供 HTTP 中间件经窄接口调用。
// 复核 token_version（旧 token 一律失效）与封禁状态。
func (s *AccountService) Authenticate(rawToken string) (*model.User, error) {
	claims, err := token.Parse(rawToken, s.secret)
	if err != nil {
		return nil, apperr.ErrInvalidToken
	}
	u, err := s.users.FindByID(claims.UID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.ErrInvalidToken
		}
		return nil, err
	}
	if u.TokenVersion != claims.TV {
		return nil, apperr.ErrInvalidToken
	}
	if u.IsBanned() {
		return nil, bannedErr(u.BanReason)
	}
	return u, nil
}

func bannedErr(reason *string) error {
	if reason == nil || *reason == "" {
		return apperr.ErrBanned
	}
	return fmt.Errorf("%w：%s", apperr.ErrBanned, *reason)
}

// normalizeEmail 去空格小写（主仓库 mailer.go NormalizeEmail 同语义）。
func normalizeEmail(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

func validEmail(s string) bool {
	return len(s) <= 254 && emailPattern.MatchString(s)
}
