package store

import (
	"testing"
	"time"

	"github.com/modern-day-team/Modernchat-sso/server/internal/model"
)

func TestEmailCode_IssueCooldown(t *testing.T) {
	db := newTestDB(t)
	s := NewEmailCodeStore(db, "test-secret-32-bytes-long-enough!!")

	code1, retry1, err := s.Issue("a@b.c", "register")
	if err != nil || code1 == "" || retry1 != 0 {
		t.Fatalf("首次发码应成功: code=%q retry=%v err=%v", code1, retry1, err)
	}
	code2, retry2, err := s.Issue("a@b.c", "register")
	if err != nil {
		t.Fatalf("第二次发码不应报错: %v", err)
	}
	if code2 != "" {
		t.Fatalf("冷却期内不应生成新码，got %q", code2)
	}
	if retry2 <= 0 {
		t.Fatalf("冷却响应应带剩余秒数，got %v", retry2)
	}
}

func TestEmailCode_VerifyOneTime(t *testing.T) {
	db := newTestDB(t)
	s := NewEmailCodeStore(db, "test-secret-32-bytes-long-enough!!")

	code, _, err := s.Issue("one@b.c", "register")
	if err != nil || code == "" {
		t.Fatalf("发码失败: %v", err)
	}
	ok, err := s.Verify("one@b.c", "register", code)
	if err != nil || !ok {
		t.Fatalf("首次校验应通过: ok=%v err=%v", ok, err)
	}
	ok, err = s.Verify("one@b.c", "register", code)
	if err != nil || ok {
		t.Fatalf("一次性：二次校验应失败: ok=%v err=%v", ok, err)
	}
}

func TestEmailCode_WrongAttemptsInvalidate(t *testing.T) {
	db := newTestDB(t)
	s := NewEmailCodeStore(db, "test-secret-32-bytes-long-enough!!")

	code, _, err := s.Issue("wrong@b.c", "register")
	if err != nil || code == "" {
		t.Fatalf("发码失败: %v", err)
	}
	// 连错 5 次（第 5 次原子作废）
	for i := 0; i < 5; i++ {
		ok, err := s.Verify("wrong@b.c", "register", "000000")
		if ok {
			t.Fatalf("错误码不应通过")
		}
		if err != nil {
			t.Fatalf("错误尝试不应报错: %v", err)
		}
	}
	// 第 5 次错误已作废：正确码也失效
	ok, err := s.Verify("wrong@b.c", "register", code)
	if err != nil || ok {
		t.Fatalf("作废后正确码也应失败: ok=%v err=%v", ok, err)
	}
}

func TestEmailCode_ReissueAfterCooldown(t *testing.T) {
	db := newTestDB(t)
	s := NewEmailCodeStore(db, "test-secret-32-bytes-long-enough!!")

	code1, _, err := s.Issue("again@b.c", "register")
	if err != nil || code1 == "" {
		t.Fatalf("首次发码失败: %v", err)
	}
	// 把 created_at 拨回冷却窗口之外，模拟时间流逝
	if err := db.Model(&model.EmailCode{}).Where("email = ?", "again@b.c").
		Update("created_at", time.Now().Add(-2*time.Minute)).Error; err != nil {
		t.Fatalf("拨时间失败: %v", err)
	}
	code2, _, err := s.Issue("again@b.c", "register")
	if err != nil || code2 == "" {
		t.Fatalf("冷却过期后应能重新发码: %v", err)
	}
	// 新码可用、旧码作废（used_at 重置但哈希已换）
	ok, err := s.Verify("again@b.c", "register", code2)
	if err != nil || !ok {
		t.Fatalf("新码应通过: ok=%v err=%v", ok, err)
	}
}
