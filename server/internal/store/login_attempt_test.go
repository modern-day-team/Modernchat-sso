package store

import (
	"testing"
	"time"

	"github.com/modern-day-team/Modernchat-sso/server/internal/model"
)

func TestLoginAttempt_LockAfterMaxFails(t *testing.T) {
	db := newTestDB(t)
	s := NewLoginAttemptStore(db)

	for i := 0; i < 5; i++ {
		if err := s.Fail("locked-user", 5, 15*time.Minute); err != nil {
			t.Fatalf("记录失败 #%d: %v", i+1, err)
		}
	}
	lockedUntil, err := s.IsLocked("locked-user")
	if err != nil {
		t.Fatalf("IsLocked: %v", err)
	}
	if !lockedUntil.After(time.Now()) {
		t.Fatalf("达阈值应进入锁定，got %v", lockedUntil)
	}
}

func TestLoginAttempt_ResetClearsCount(t *testing.T) {
	db := newTestDB(t)
	s := NewLoginAttemptStore(db)

	_ = s.Fail("reset-user", 5, 15*time.Minute)
	_ = s.Fail("reset-user", 5, 15*time.Minute)
	if err := s.Reset("reset-user"); err != nil {
		t.Fatalf("Reset: %v", err)
	}
	lockedUntil, err := s.IsLocked("reset-user")
	if err != nil {
		t.Fatalf("IsLocked: %v", err)
	}
	if !lockedUntil.IsZero() {
		t.Fatalf("清零后不应锁定，got %v", lockedUntil)
	}
	// 清零后再次连错需重新累计：2 次不应锁定
	_ = s.Fail("reset-user", 5, 15*time.Minute)
	_ = s.Fail("reset-user", 5, 15*time.Minute)
	lockedUntil, _ = s.IsLocked("reset-user")
	if lockedUntil.After(time.Now()) {
		t.Fatalf("清零后 2 次失败不应锁定")
	}
}

func TestLoginAttempt_LockExpiredThenRecount(t *testing.T) {
	db := newTestDB(t)
	s := NewLoginAttemptStore(db)

	for i := 0; i < 5; i++ {
		_ = s.Fail("expire-user", 5, 15*time.Minute)
	}
	// 锁定期间计数已清零：把 locked_until 拨到过去模拟锁过期
	if err := db.Model(&model.LoginAttempt{}).Where("username = ?", "expire-user").
		Update("locked_until", time.Now().Add(-time.Minute)).Error; err != nil {
		t.Fatalf("拨时间失败: %v", err)
	}
	lockedUntil, err := s.IsLocked("expire-user")
	if err != nil {
		t.Fatalf("IsLocked: %v", err)
	}
	if lockedUntil.After(time.Now()) {
		t.Fatalf("锁过期后不应再锁定")
	}
}
