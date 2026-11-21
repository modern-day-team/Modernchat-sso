package store

import (
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/modern-day-team/Modernchat-sso/server/internal/model"
)

func passwordResetForTest(userID uint, token string) *model.PasswordReset {
	return &model.PasswordReset{
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}
}

func TestPasswordReset_UseSingleConsume(t *testing.T) {
	db := newTestDB(t)
	s := NewPasswordResetStore(db)
	u := newUser(t, db, "reset-consume")

	pr := passwordResetForTest(u.ID, "1111111111111111111111111111111111111111111111111111111111111111")
	if err := s.Create(pr); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := s.Valid(pr.Token)
	if err != nil || got == nil {
		t.Fatalf("Valid 应找到有效令牌: pr=%v err=%v", got, err)
	}

	// 第一次消费：成功并在事务内执行回调（模拟写新密码 + tv++）
	applied := false
	err = s.Use(got, func(tx *gorm.DB) error {
		applied = true
		return nil
	})
	if err != nil || !applied {
		t.Fatalf("首次消费应成功并执行回调: applied=%v err=%v", applied, err)
	}

	// 第二次消费：ErrResetTokenUsed
	err = s.Use(got, func(tx *gorm.DB) error {
		t.Fatal("已消费的令牌不应再次执行回调")
		return nil
	})
	if !errors.Is(err, ErrResetTokenUsed) {
		t.Fatalf("重复消费应返回 ErrResetTokenUsed，got %v", err)
	}

	// Valid 不再命中
	got, err = s.Valid(pr.Token)
	if err != nil || got != nil {
		t.Fatalf("消费后 Valid 应返回 nil: pr=%v err=%v", got, err)
	}
}

func TestPasswordReset_ValidRejectsExpired(t *testing.T) {
	db := newTestDB(t)
	s := NewPasswordResetStore(db)
	u := newUser(t, db, "reset-expired")

	pr := passwordResetForTest(u.ID, "2222222222222222222222222222222222222222222222222222222222222222")
	pr.ExpiresAt = time.Now().Add(-time.Minute)
	if err := s.Create(pr); err != nil {
		t.Fatalf("Create: %v", err)
	}
	got, err := s.Valid(pr.Token)
	if err != nil || got != nil {
		t.Fatalf("过期令牌应视为无效: pr=%v err=%v", got, err)
	}
}
