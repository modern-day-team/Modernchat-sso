// Package service 业务层：账号/认证业务规则唯一归属地。
// 只依赖 store、token、apperr，不感知 HTTP（gin）；错误一律返回 apperr 哨兵。
package service

import (
	"cmp"
	"errors"
	"fmt"
	"log/slog"
	"regexp"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/modern-day-team/Modernchat-sso/server/internal/apperr"
	"github.com/modern-day-team/Modernchat-sso/server/internal/model"
	"github.com/modern-day-team/Modernchat-sso/server/internal/store"
	"github.com/modern-day-team/Modernchat-sso/server/internal/token"
)

// 头像默认取色板（与主仓库 TDesign 色系一致）。
var avatarPalette = []string{"#0052d9", "#2ba471", "#d54941", "#ed7b2f", "#ca8d04"}

// AccountService 账号业务。
type AccountService struct {
	users  *store.UserStore
	secret string // JWT 签名密钥（与主仓库共享同一值即 token 互认）
}

// NewAccountService 构造。
func NewAccountService(users *store.UserStore, secret string) *AccountService {
	return &AccountService{users: users, secret: secret}
}

// usernamePattern 用户名规则与主仓库一致：3-20 位，仅字母/数字/下划线/连字符（规格 §4）。
var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,20}$`)

// Register 注册：参数校验、唯一性、首个注册者为 admin（沿用主仓库约定）、默认头像色。
func (s *AccountService) Register(username, password, nickname string, email *string) (*model.User, error) {
	if !usernamePattern.MatchString(username) || len(password) < 6 || len(password) > 72 {
		return nil, apperr.ErrInvalidInput
	}
	if email != nil && !validEmail(*email) {
		return nil, apperr.ErrInvalidEmail
	}
	if _, err := s.users.FindByUsername(username); err == nil {
		return nil, apperr.ErrUsernameTaken
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if email != nil {
		if _, err := s.users.FindByEmail(*email); err == nil {
			return nil, apperr.ErrEmailTaken
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	count, err := s.users.Count()
	if err != nil {
		return nil, err
	}
	u := model.User{
		Username:     username,
		PasswordHash: string(hash),
		Nickname:     cmp.Or(nickname, username),
		AvatarColor:  avatarPalette[int(count)%len(avatarPalette)],
		Email:        email,
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

// Login 登录：校验凭证与封禁，签发 token。
// 家族规则：登录失败只记用户名与 IP，绝不记密码/令牌；账号不存在与密码错误同一错误。
func (s *AccountService) Login(username, password, ip string) (string, *model.User, error) {
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
		return "", nil, apperr.ErrInvalidCredentials
	}
	if u.IsBanned() {
		return "", nil, bannedErr(u.BanReason)
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

// 邮箱规则与主仓库一致（规格 §4）：^[^\s@]+@[^\s@]+\.[^\s@]+$，≤254 字符
var emailPattern = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

func validEmail(s string) bool {
	return len(s) <= 254 && emailPattern.MatchString(s)
}
