package model

import "time"

// LoginAttempt 登录失败锁定：按用户名计数（非 IP），连错达阈值锁定一段时间。
type LoginAttempt struct {
	ID          uint       `gorm:"primaryKey" json:"-"`
	Username    string     `gorm:"size:20;uniqueIndex" json:"-"`
	FailedCount int        `gorm:"not null;default:0" json:"-"`
	LockedUntil *time.Time `gorm:"type:datetime(3)" json:"-"`
	UpdatedAt   time.Time  `gorm:"type:datetime(3)" json:"-"`
}

// TableName 表名。
func (LoginAttempt) TableName() string { return "login_attempts" }
