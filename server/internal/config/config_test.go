package config

import (
	"os"
	"strings"
	"testing"
)

// envKeys Load 读取的全部环境变量。测试期间统一设为安全占位值，
// 避免 godotenv.Load() 从磁盘 .env 注入真实配置（godotenv 对已存在的键不覆盖）。
var envKeys = []string{
	"PORT", "DB_DSN", "JWT_SECRET", "COOKIE_SECURE", "PUBLIC_URL", "WEB_UPDATE",
	"RELEASE_API_URL", "RELEASE_DL_URL", "RELEASE_REPO_GITHUB", "GITHUB_TOKEN",
	"MAIL_PROVIDER", "SMTP_HOST", "SMTP_PORT", "SMTP_USER", "SMTP_PASS", "SMTP_FROM",
	"TENCENT_SES_SECRET_ID", "TENCENT_SES_SECRET_KEY", "TENCENT_SES_REGION", "TENCENT_SES_FROM", "TENCENT_SES_TEMPLATE_ID",
}

// isolateEnv 把全部相关环境变量置空后按 set 设置，测试结束还原。
// 注意：本包测试不可并行（环境变量是进程级全局状态）。
func isolateEnv(t *testing.T, set map[string]string) {
	t.Helper()
	saved := make(map[string]*string, len(envKeys))
	for _, k := range envKeys {
		if v, ok := os.LookupEnv(k); ok {
			vv := v
			saved[k] = &vv
		} else {
			saved[k] = nil
		}
		_ = os.Setenv(k, "")
	}
	for k, v := range set {
		_ = os.Setenv(k, v)
	}
	t.Cleanup(func() {
		for k, v := range saved {
			if v == nil {
				_ = os.Unsetenv(k)
			} else {
				_ = os.Setenv(k, *v)
			}
		}
	})
}

const validSecret = "test-jwt-secret-0123456789abcdef-0123456789abcdef"

// TestLoadRequiresJWTSecret 不提供弱默认：空或短密钥必须拒绝启动。
func TestLoadRequiresJWTSecret(t *testing.T) {
	isolateEnv(t, map[string]string{"JWT_SECRET": ""})
	if _, err := Load(); err == nil {
		t.Fatal("JWT_SECRET 为空应报错")
	}
	isolateEnv(t, map[string]string{"JWT_SECRET": "too-short"})
	if _, err := Load(); err == nil {
		t.Fatal("JWT_SECRET 不足 32 字节应报错")
	}
}

// TestLoadDefaults 合法密钥下各字段取开发默认值。
func TestLoadDefaults(t *testing.T) {
	isolateEnv(t, map[string]string{"JWT_SECRET": validSecret})
	c, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.Port != "8090" || c.PublicURL != "http://localhost:5175" {
		t.Fatalf("默认端口/公开地址不符: %+v", c)
	}
	if !strings.Contains(c.DBDSN, "3307") || !strings.Contains(c.DBDSN, "modernchat_sso") {
		t.Fatalf("默认 DSN 应指向本地 SSO 库: %q", c.DBDSN)
	}
	if c.MailProvider != "console" || c.CookieSecure || !c.WebUpdate {
		t.Fatalf("默认开关不符: %+v", c)
	}
}

// TestLoadEnvOverride 环境变量覆盖默认值，且布尔/数值按项目组口径解析。
func TestLoadEnvOverride(t *testing.T) {
	isolateEnv(t, map[string]string{
		"JWT_SECRET":              validSecret,
		"PORT":                    "9000",
		"PUBLIC_URL":              "https://sso.example.com",
		"COOKIE_SECURE":           "true",
		"WEB_UPDATE":              "false",
		"MAIL_PROVIDER":           "tencent_ses,smtp",
		"TENCENT_SES_TEMPLATE_ID": "12345",
		"DB_DSN":                  "root:pw@tcp(db:3306)/sso",
	})
	c, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.Port != "9000" || c.PublicURL != "https://sso.example.com" {
		t.Fatalf("覆盖失败: %+v", c)
	}
	if !c.CookieSecure || c.WebUpdate {
		t.Fatalf("布尔解析失败: secure=%v webUpdate=%v", c.CookieSecure, c.WebUpdate)
	}
	if c.MailProvider != "tencent_ses,smtp" || c.TencentSESTemplateID != 12345 {
		t.Fatalf("多厂商/模板 ID 解析失败: %+v", c)
	}
	if c.DBDSN != "root:pw@tcp(db:3306)/sso" {
		t.Fatalf("DSN 覆盖失败: %q", c.DBDSN)
	}
}
