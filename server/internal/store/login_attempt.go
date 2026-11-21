package store

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/modern-day-team/Modernchat-sso/server/internal/model"
)

// LoginAttemptStore login_attempts 表访问：按用户名计登录失败（非 IP），
// 连错达阈值锁定一段时间；凭据正确即清零。
type LoginAttemptStore struct{ db *gorm.DB }

// NewLoginAttemptStore 构造。
func NewLoginAttemptStore(db *gorm.DB) *LoginAttemptStore { return &LoginAttemptStore{db: db} }

// IsLocked 返回锁定截止时间；未锁定返回零值。
func (s *LoginAttemptStore) IsLocked(username string) (time.Time, error) {
	var a model.LoginAttempt
	err := s.db.Where("username = ?", username).First(&a).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return time.Time{}, nil
		}
		return time.Time{}, err
	}
	if a.LockedUntil != nil && a.LockedUntil.After(time.Now()) {
		return *a.LockedUntil, nil
	}
	return time.Time{}, nil
}

// Fail 记一次失败：failed_count+1（upsert）；达阈值时置 locked_until 并清零计数。
// 语义与主仓库一致：锁定期间计数清零，锁定到期后重新累计。
func (s *LoginAttemptStore) Fail(username string, maxAttempts int, lockFor time.Duration) error {
	now := time.Now()
	if err := s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "username"}},
		DoUpdates: clause.Assignments(map[string]any{"failed_count": gorm.Expr("login_attempts.failed_count + 1")}),
	}).Create(&model.LoginAttempt{Username: username, FailedCount: 1, UpdatedAt: now}).Error; err != nil {
		return err
	}
	var a model.LoginAttempt
	if err := s.db.Where("username = ?", username).First(&a).Error; err != nil {
		return err
	}
	if a.FailedCount >= maxAttempts {
		until := now.Add(lockFor)
		return s.db.Model(&model.LoginAttempt{}).Where("username = ?", username).
			Updates(map[string]any{"locked_until": until, "failed_count": 0}).Error
	}
	return nil
}

// Reset 登录成功后清零。
func (s *LoginAttemptStore) Reset(username string) error {
	return s.db.Model(&model.LoginAttempt{}).Where("username = ?", username).
		Updates(map[string]any{"failed_count": 0, "locked_until": nil}).Error
}
