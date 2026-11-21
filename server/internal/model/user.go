// Package model GORM 模型与领域类型。
package model

import "time"

// User 用户。认证相关字段与主仓库 Modernchat-new 保持同构，便于账号体系平滑接管。
type User struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	Username     string `gorm:"size:20;uniqueIndex" json:"username"`
	PasswordHash string `gorm:"size:60" json:"-"`
	Nickname     string `gorm:"size:20" json:"nickname"`
	AvatarColor  string `gorm:"size:16" json:"avatarColor"`

	Email         *string    `gorm:"size:254;uniqueIndex" json:"email,omitempty"`
	EmailVerified bool       `json:"-"`
	Role          string     `gorm:"size:10;default:user" json:"role"` // user | admin
	TokenVersion  int64      `gorm:"not null;default:1" json:"-"`      // 登出/改密时 +1，旧 token 全部失效
	BannedUntil   *time.Time `json:"bannedUntil,omitempty"`
	BanReason     *string    `gorm:"size:255" json:"banReason,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"-"`
}

// IsBanned 当前是否处于封禁期。
func (u *User) IsBanned() bool {
	return u.BannedUntil != nil && u.BannedUntil.After(time.Now())
}
