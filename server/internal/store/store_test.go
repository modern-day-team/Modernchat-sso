package store

import (
	"os"
	"testing"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/modern-day-team/Modernchat-sso/server/internal/model"
)

// newTestDB 连接 CI/本地注入的 MySQL（DB_DSN 环境变量）；不可用则跳过。
// ⚠️ 主仓库教训：连不上库会 Skip 且 go test 仍报 ok——CI 已配 MySQL 服务，本地无库时
// 这些测试不产生覆盖，属已知状态（见 AGENTS.md）。
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		t.Skip("DB_DSN 未设置，跳过存储层测试（CI 已配置 MySQL 服务）")
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("数据库连接失败，跳过存储层测试：%v", err)
	}
	if err := db.AutoMigrate(
		&model.User{}, &model.SystemSetting{}, &model.EmailCode{},
		&model.LoginAttempt{}, &model.PasswordReset{},
	); err != nil {
		t.Fatalf("AutoMigrate: %v", err)
	}
	cleanup := func() {
		db.Exec("DELETE FROM password_resets")
		db.Exec("DELETE FROM email_codes")
		db.Exec("DELETE FROM login_attempts")
		db.Exec("DELETE FROM system_settings")
		db.Exec("DELETE FROM users")
	}
	cleanup()
	t.Cleanup(cleanup)
	return db
}

// newUser 建一个测试用户（唯一用户名避免冲突）；PasswordHash 用真实 bcrypt（60 字符，匹配列宽）。
func newUser(t *testing.T, db *gorm.DB, username string) model.User {
	t.Helper()
	h, err := bcrypt.GenerateFromPassword([]byte("test-password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("生成 bcrypt 失败: %v", err)
	}
	u := model.User{
		Username:     username,
		PasswordHash: string(h),
		Nickname:     username,
	}
	if err := db.Create(&u).Error; err != nil {
		t.Fatalf("创建测试用户失败: %v", err)
	}
	return u
}
