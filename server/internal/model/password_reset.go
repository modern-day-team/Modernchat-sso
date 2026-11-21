package model

import "time"

// PasswordReset 找回密码令牌：crypto/rand 32 字节 hex（64 字符），15 分钟有效，一次性。
type PasswordReset struct {
	ID        uint      `gorm:"primaryKey" json:"-"`
	UserID    uint      `json:"-"`
	Token     string    `gorm:"size:64;uniqueIndex" json:"-"`
	ExpiresAt time.Time `json:"-"`
	Used      bool      `json:"-"`
	CreatedAt time.Time `json:"-"`
}

// TableName 表名。
func (PasswordReset) TableName() string { return "password_resets" }
