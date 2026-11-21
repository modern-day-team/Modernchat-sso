package model

import "time"

// EmailCode 邮箱验证码：明文不落库，只存 HMAC-SHA256 哈希；
// (Email, Purpose) 复合唯一索引保证每个用途单行，过期/作废靠 UsedAt + Attempts。
type EmailCode struct {
	ID        uint       `gorm:"primaryKey" json:"-"`
	Email     string     `gorm:"size:254;not null;uniqueIndex:uk_email_codes_email_purpose,priority:1" json:"-"`
	CodeHash  string     `gorm:"size:64;not null" json:"-"`
	Purpose   string     `gorm:"size:16;not null;default:register;uniqueIndex:uk_email_codes_email_purpose,priority:2" json:"-"`
	Attempts  int        `gorm:"not null;default:0" json:"-"`
	ExpiresAt time.Time  `gorm:"type:datetime(3);not null" json:"-"`
	UsedAt    *time.Time `gorm:"type:datetime(3)" json:"-"`
	CreatedAt time.Time  `gorm:"type:datetime(3)" json:"-"`
}

// TableName 表名。
func (EmailCode) TableName() string { return "email_codes" }
