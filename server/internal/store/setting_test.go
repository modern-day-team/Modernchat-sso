package store

import "testing"

// TestSettingStore_SetGet 覆盖 key/value 往返。key 是 MySQL 8.0 保留字，
// 裸写 Where("key = ?") 会语法错误并被 Get 当成"不存在"——本用例专门防回归。
func TestSettingStore_SetGet(t *testing.T) {
	db := newTestDB(t)
	s := NewSettingStore(db)

	if err := s.Set("test.setting_key", "hello"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	v, ok, err := s.Get("test.setting_key")
	if err != nil {
		t.Fatalf("Get 不应报错: %v", err)
	}
	if !ok || v != "hello" {
		t.Fatalf("Get 往返失败: ok=%v value=%q", ok, v)
	}
	if err := s.Set("test.setting_key", "world"); err != nil {
		t.Fatalf("覆盖写 Set: %v", err)
	}
	v, ok, _ = s.Get("test.setting_key")
	if !ok || v != "world" {
		t.Fatalf("覆盖写后 Get: ok=%v value=%q", ok, v)
	}
}

// TestSettingStore_OpsRoundTrip 覆盖整组站点开关的读写，确保站长设置保存后能读回。
func TestSettingStore_OpsRoundTrip(t *testing.T) {
	db := newTestDB(t)
	s := NewSettingStore(db)

	want := Ops{
		AllowRegister:      false,
		RequireEmailVerify: false,
		RequireInviteCode:  true,
		MaintenanceMode:    true,
		MaintenanceMessage: "站点维护中",
		LoginMaxAttempts:   0,
		LoginLockMinutes:   30,
	}
	if err := s.UpdateOps(want); err != nil {
		t.Fatalf("UpdateOps: %v", err)
	}
	got := s.Ops()
	if got != want {
		t.Fatalf("Ops 往返不一致：want=%+v got=%+v", want, got)
	}
}
