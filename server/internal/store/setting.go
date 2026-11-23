package store

import (
	"errors"
	"strconv"

	"gorm.io/gorm"

	"github.com/modern-day-team/Modernchat-sso/server/internal/model"
)

// 站点开关的键名。
const (
	KeyAllowRegister      = "ops.allow_register"
	KeyRequireEmailVerify = "ops.require_email_verify"
	KeyRequireInviteCode  = "ops.require_invite_code"
	KeyMaintenanceMode    = "ops.maintenance_mode"
	KeyMaintenanceMsg     = "ops.maintenance_message"
	KeyLoginMaxAttempts   = "ops.login_max_attempts"
	KeyLoginLockMinutes   = "ops.login_lock_minutes"
)

// Ops 站点运行开关。默认值与主仓库 system_settings 默认值一致（store/setting.go:249-269）。
type Ops struct {
	AllowRegister      bool
	RequireEmailVerify bool
	RequireInviteCode  bool
	MaintenanceMode    bool
	MaintenanceMessage string
	LoginMaxAttempts   int // 0 = 关闭锁定；范围 2-20
	LoginLockMinutes   int // 范围 1-1440
}

// defaultOps 默认开关（首次部署、表中无行时的取值）。
var defaultOps = Ops{
	AllowRegister:      true,
	RequireEmailVerify: true,
	LoginMaxAttempts:   5,
	LoginLockMinutes:   15,
}

// DefaultOps 默认开关（对外暴露，供 handler 在 SettingStore 缺失时兜底）。
func DefaultOps() Ops { return defaultOps }

// SettingStore system_settings 表访问。
type SettingStore struct{ db *gorm.DB }

// NewSettingStore 构造。
func NewSettingStore(db *gorm.DB) *SettingStore { return &SettingStore{db: db} }

// Get 读单个键；不存在返回 (空串, false, nil)。
func (s *SettingStore) Get(key string) (string, bool, error) {
	var v model.SystemSetting
	// 用结构体条件让 GORM 转义列名：key 是 MySQL 8.0 保留字，
	// 裸写 Where("key = ?") 会生成语法错误并被 Ops() 当成"键不存在"静默回退默认值。
	err := s.db.Where(&model.SystemSetting{Key: key}).First(&v).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", false, nil
		}
		return "", false, err
	}
	return v.Value, true, nil
}

// Set 写入单个键（upsert）。
func (s *SettingStore) Set(key, value string) error {
	return s.db.Save(&model.SystemSetting{Key: key, Value: value}).Error
}

// Ops 读取运行开关；表中缺失的键回退默认值。
func (s *SettingStore) Ops() Ops {
	ops := defaultOps
	if s.db == nil {
		return ops
	}
	if v, ok := s.getBool(KeyAllowRegister); ok {
		ops.AllowRegister = v
	}
	if v, ok := s.getBool(KeyRequireEmailVerify); ok {
		ops.RequireEmailVerify = v
	}
	if v, ok := s.getBool(KeyRequireInviteCode); ok {
		ops.RequireInviteCode = v
	}
	if v, ok := s.getBool(KeyMaintenanceMode); ok {
		ops.MaintenanceMode = v
	}
	if v, ok, _ := s.Get(KeyMaintenanceMsg); ok {
		ops.MaintenanceMessage = v
	}
	if v, ok := s.getInt(KeyLoginMaxAttempts); ok {
		ops.LoginMaxAttempts = v
	}
	if v, ok := s.getInt(KeyLoginLockMinutes); ok {
		ops.LoginLockMinutes = v
	}
	return ops
}

// UpdateOps 覆盖运行开关（站长侧）。
func (s *SettingStore) UpdateOps(ops Ops) error {
	items := map[string]string{
		KeyAllowRegister:      strconv.FormatBool(ops.AllowRegister),
		KeyRequireEmailVerify: strconv.FormatBool(ops.RequireEmailVerify),
		KeyRequireInviteCode:  strconv.FormatBool(ops.RequireInviteCode),
		KeyMaintenanceMode:    strconv.FormatBool(ops.MaintenanceMode),
		KeyMaintenanceMsg:     ops.MaintenanceMessage,
		KeyLoginMaxAttempts:   strconv.Itoa(ops.LoginMaxAttempts),
		KeyLoginLockMinutes:   strconv.Itoa(ops.LoginLockMinutes),
	}
	for k, v := range items {
		if err := s.Set(k, v); err != nil {
			return err
		}
	}
	return nil
}

func (s *SettingStore) getBool(key string) (bool, bool) {
	v, ok, err := s.Get(key)
	if err != nil || !ok {
		return false, false
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return false, false
	}
	return b, true
}

func (s *SettingStore) getInt(key string) (int, bool) {
	v, ok, err := s.Get(key)
	if err != nil || !ok {
		return 0, false
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, false
	}
	return n, true
}
