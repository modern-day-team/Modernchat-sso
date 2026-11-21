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
// 冷却期内返回 ("", 剩余秒数, nil)，调用方不下发邮件（防轰炸）。
// 单行 (email,purpose)：冷却过期后原行重置（清 attempts/used_at）；并发首发视为冷却。
func (s *EmailCodeStore) Issue(email, purpose string) (string, time.Duration, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", 0, err
	}
	code := fmt.Sprintf("%06d", n.Int64())
	codeHash := s.hash(email, purpose, code)
	now := time.Now()
	var retryAfter time.Duration

	err = s.db.Transaction(func(tx *gorm.DB) error {
		var last model.EmailCode
		rowErr := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("email = ? AND purpose = ?", email, purpose).Limit(1).First(&last).Error
		switch {
		case errors.Is(rowErr, gorm.ErrRecordNotFound):
			cr := tx.Create(&model.EmailCode{
				Email: email, CodeHash: codeHash, Purpose: purpose,
				ExpiresAt: now.Add(EmailCodeTTL),
			})
			if cr.Error != nil {
				return cr.Error
			}
			// 并发首发撞唯一索引（INSERT IGNORE 生效）：按冷却处理
			if cr.RowsAffected == 0 {
				retryAfter = EmailCodeCooldown
			}
			return nil
		case rowErr != nil:
			return rowErr
		}
		if elapsed := now.Sub(last.CreatedAt); elapsed < EmailCodeCooldown {
			retryAfter = EmailCodeCooldown - elapsed
			return nil
		}
		// 冷却过期：原行重置（清错误计数与已用标记）
		return tx.Model(&last).Updates(map[string]any{
			"code_hash":  codeHash,
			"attempts":   0,
			"used_at":    nil,
			"expires_at": now.Add(EmailCodeTTL),
			"created_at": now,
		}).Error
	})
	if err != nil {
		return "", 0, err
	}
	if code == "" {
		return "", retryAfter, nil
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
