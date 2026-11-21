package store

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/modern-day-team/Modernchat-sso/server/internal/model"
)

// 邮箱验证码参数（与主仓库 Modernchat-new 完全一致）。
const (
	EmailCodeTTL         = 10 * time.Minute
	EmailCodeCooldown    = 60 * time.Second
	emailCodeMaxAttempts = 5
	emailCodePurgeAfter  = 24 * time.Hour
)

// EmailCodeStore email_codes 表访问：明文验证码不落库，只存 HMAC-SHA256 哈希。
// 哈希密钥为 JWT_SECRET 原字节——换密钥后旧码自动失效。
type EmailCodeStore struct {
	db     *gorm.DB
	secret []byte
}

// NewEmailCodeStore 构造。
func NewEmailCodeStore(db *gorm.DB, secret string) *EmailCodeStore {
	return &EmailCodeStore{db: db, secret: []byte(secret)}
}

func (s *EmailCodeStore) hash(email, purpose, code string) string {
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(email + "|" + purpose + "|" + code))
	return hex.EncodeToString(mac.Sum(nil))
}

// Issue 生成 6 位验证码并落哈希。返回 (code, retryAfter, err)：
// 冷却期内返回 ("", 60, nil)，调用方不下发邮件（防轰炸）。
// 冷却检查用 SQL 级时间比较（避免回读 created_at 的时区/精度坑，CI 实测踩坑后重构）；
// 单行 (email,purpose) 用 ON DUPLICATE KEY UPDATE 原子 upsert（并发安全，无需 FOR UPDATE）。
func (s *EmailCodeStore) Issue(email, purpose string) (string, time.Duration, error) {
	now := time.Now()
	var recent int64
	if err := s.db.Model(&model.EmailCode{}).
		Where("email = ? AND purpose = ? AND created_at > ?", email, purpose, now.Add(-EmailCodeCooldown)).
		Count(&recent).Error; err != nil {
		return "", 0, err
	}
	if recent > 0 {
		return "", EmailCodeCooldown, nil
	}

	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", 0, err
	}
	code := fmt.Sprintf("%06d", n.Int64())
	codeHash := s.hash(email, purpose, code)

	// 首发走 INSERT，冷却过期/并发撞索引走 UPDATE（原子，均为单行）
	cr := s.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "email"}, {Name: "purpose"}},
		DoUpdates: clause.Assignments(map[string]any{
			"code_hash":  codeHash,
			"attempts":   0,
			"used_at":    nil,
			"expires_at": now.Add(EmailCodeTTL),
			"created_at": now,
		}),
	}).Create(&model.EmailCode{
		Email:     email,
		CodeHash:  codeHash,
		Purpose:   purpose,
		ExpiresAt: now.Add(EmailCodeTTL),
		CreatedAt: now,
	})
	if cr.Error != nil {
		return "", 0, cr.Error
	}
	if cr.RowsAffected == 0 {
		// 理论不可达（哈希每次变化）：并发撞冷却兜底
		return "", EmailCodeCooldown, nil
	}
	// 顺手清理 24h 前的旧行
	s.db.Where("created_at < ?", now.Add(-emailCodePurgeAfter)).Delete(&model.EmailCode{})
	return code, 0, nil
}

// Verify 校验验证码：一次性（条件 UPDATE 消费）、常数时间比较、
// 错误计数原子 +1 且第 5 次错误在同一个 UPDATE 里作废。
func (s *EmailCodeStore) Verify(email, purpose, code string) (bool, error) {
	want := s.hash(email, purpose, code)
	now := time.Now()
	var row model.EmailCode
	err := s.db.Where("email = ? AND purpose = ? AND used_at IS NULL AND expires_at > ?", email, purpose, now).
		Order("id DESC").Limit(1).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if subtle.ConstantTimeCompare([]byte(row.CodeHash), []byte(want)) != 1 {
		err := s.db.Model(&model.EmailCode{}).Where("id = ?", row.ID).Updates(map[string]any{
			"attempts": gorm.Expr("attempts + 1"),
			"used_at":  gorm.Expr("CASE WHEN attempts + 1 >= ? THEN ? ELSE used_at END", emailCodeMaxAttempts, now),
		}).Error
		if err != nil {
			return false, err
		}
		return false, nil
	}
	res := s.db.Model(&model.EmailCode{}).
		Where("id = ? AND used_at IS NULL", row.ID).Update("used_at", now)
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}
