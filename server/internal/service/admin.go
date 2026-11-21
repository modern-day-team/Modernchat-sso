package service

import (
	"crypto/rand"
	"errors"
	"math/big"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/modern-day-team/Modernchat-sso/server/internal/apperr"
	"github.com/modern-day-team/Modernchat-sso/server/internal/model"
	"github.com/modern-day-team/Modernchat-sso/server/internal/store"
)

// resetPasswordAlphabet 站长重置密码的随机字母表：排除易混淆字符（i/l/1/0/o，主仓库同款）。
const resetPasswordAlphabet = "abcdefghjkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"

// AdminService 站长账号管理业务（约束与副作用在此，handler 只做 HTTP 映射）。
type AdminService struct {
	users         *store.UserStore
	loginAttempts *store.LoginAttemptStore
}

// NewAdminService 构造。
func NewAdminService(users *store.UserStore, loginAttempts *store.LoginAttemptStore) *AdminService {
	return &AdminService{users: users, loginAttempts: loginAttempts}
}

// List 分页用户列表（page/pageSize 归一在 store 层）。
func (s *AdminService) List(page, pageSize int, kw, role, status string) ([]model.User, int64, error) {
	return s.users.List(store.UserListQuery{
		Page: page, PageSize: pageSize, Kw: kw, Role: role, Status: status,
	})
}

// userByID 按 id 取用户；不存在返回 404 语义错误。
func (s *AdminService) userByID(id uint) (*model.User, error) {
	u, err := s.users.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.ErrUserNotFound
		}
		return nil, err
	}
	return u, nil
}

// Ban 封禁：days 0-3650（0=永久 +100 年）；封禁 admin 角色→403；tv bump 全端失效。
func (s *AdminService) Ban(id uint, days int, reason string) error {
	if days < 0 || days > 3650 {
		return apperr.ErrInvalidDays
	}
	u, err := s.userByID(id)
	if err != nil {
		return err
	}
	if u.Role == "admin" {
		return apperr.ErrCannotBanAdmin
	}
	var until time.Time
	if days == 0 {
		until = time.Now().AddDate(100, 0, 0)
	} else {
		until = time.Now().AddDate(0, 0, days)
	}
	u.BannedUntil = &until
	if reason != "" {
		u.BanReason = &reason
	} else {
		u.BanReason = nil // 清掉上次封禁残留的原因
	}
	u.TokenVersion++
	return s.users.Save(u)
}

// Unban 解封：清 BannedUntil/BanReason；不 bump tv（被禁用户本就无有效 token，主仓库同款）。
func (s *AdminService) Unban(id uint) error {
	u, err := s.userByID(id)
	if err != nil {
		return err
	}
	u.BannedUntil = nil
	u.BanReason = nil
	return s.users.Save(u)
}

// ResetPassword 生成 12 位随机密码（排除易混字符），tv bump 全端失效；返回新密码供站长转达。
func (s *AdminService) ResetPassword(id uint) (string, error) {
	u, err := s.userByID(id)
	if err != nil {
		return "", err
	}
	b := make([]byte, 12)
	for i := range b {
		k, err := rand.Int(rand.Reader, big.NewInt(int64(len(resetPasswordAlphabet))))
		if err != nil {
			return "", err
		}
		b[i] = resetPasswordAlphabet[k.Int64()]
	}
	newPassword := string(b)
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	u.PasswordHash = string(hash)
	u.TokenVersion++
	if err := s.users.Save(u); err != nil {
		return "", err
	}
	return newPassword, nil
}

// SetRole 角色变更（admin/user）；降职最后站长→403。
// 不 bump tv——Auth 每请求查库，角色变更即时生效（主仓库注释同款）。
func (s *AdminService) SetRole(id uint, role string) error {
	if role != "user" && role != "admin" {
		return apperr.ErrInvalidRole
	}
	err := s.users.SetRole(id, role)
	if errors.Is(err, store.ErrLastAdmin) {
		return apperr.ErrLastAdmin
	}
	return err
}
