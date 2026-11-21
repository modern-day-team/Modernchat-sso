package store

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/modern-day-team/Modernchat-sso/server/internal/model"
)

// ErrResetTokenUsed 重置令牌已被消费（并发重复提交或过期后的重放）。
var ErrResetTokenUsed = errors.New("password reset token already used")

// PasswordResetStore password_resets 表访问。
type PasswordResetStore struct{ db *gorm.DB }

// NewPasswordResetStore 构造。
func NewPasswordResetStore(db *gorm.DB) *PasswordResetStore {
	return &PasswordResetStore{db: db}
}

// Create 落一条新令牌。
func (s *PasswordResetStore) Create(pr *model.PasswordReset) error { return s.db.Create(pr).Error }

// Valid 查询有效令牌；无效（不存在/已用/过期）返回 (nil, nil)。
func (s *PasswordResetStore) Valid(token string) (*model.PasswordReset, error) {
	var pr model.PasswordReset
	err := s.db.Where("token = ? AND used = ? AND expires_at > ?", token, false, time.Now()).
		First(&pr).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &pr, nil
}

// Use 原子消费令牌：条件 UPDATE（used=true）成功后才在同一个事务里执行 fn
// （写新密码 + token_version++）。消费失败返回 ErrResetTokenUsed，fn 不执行——
// 防 TOCTOU：两个并发重置只有一个能赢（主仓库已修的缺陷模式）。
func (s *PasswordResetStore) Use(pr *model.PasswordReset, fn func(tx *gorm.DB) error) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&model.PasswordReset{}).
			Where("id = ? AND used = ? AND expires_at > ?", pr.ID, false, time.Now()).
			Update("used", true)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrResetTokenUsed
		}
		return fn(tx)
	})
}
