// Package config 配置加载：.env 可选，环境变量优先。
package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config 服务运行配置。
type Config struct {
	Port         string // HTTP 监听端口
	DBDSN        string // MySQL 连接串
	JWTSecret    string // JWT 签名密钥
	CookieSecure bool   // mc_auth cookie 是否带 Secure（生产必须 true）

	// 在线更新（站长侧）：WEB_UPDATE=false 整体关闭；可用性另需 systemd 模板单元部署
	WebUpdate         bool
	ReleaseAPIURL     string // RELEASE_API_URL 显式 releases/latest API 地址
	ReleaseDLURL      string // RELEASE_DL_URL 显式下载前缀（…/releases/download）
	ReleaseRepoGitHub string // RELEASE_REPO_GITHUB 托管仓库（owner/repo）；API/DL 未配置时派生
	GitHubToken       string // GITHUB_TOKEN 可选：防匿名 API 限流，支持私有仓库检查

	// 邮件发送（邮箱验证用）：通道 + SMTP 参数；厂商 API 预留同一 Mailer 接口接入
	MailProvider string // console（默认，仅本地开发）| smtp | tencent_ses
	SMTPHost     string
	SMTPPort     string // 465 隐式 TLS；587/25 STARTTLS
	SMTPUser     string
	SMTPPass     string
	SMTPFrom     string

	// 腾讯云 SES（MAIL_PROVIDER=tencent_ses 时必填）：凭据走环境变量，严禁入库
	TencentSESSecretID   string
	TencentSESSecretKey  string
	TencentSESRegion     string // 缺省 ap-guangzhou
	TencentSESFrom       string // 发件地址（可带别名），须已在 SES 控制台验证
	TencentSESTemplateID uint64 // 正文为 {{content}} 的已过审模板 ID
}

// Load 读取配置；.env 不存在时静默跳过（生产用环境变量注入）。
func Load() (Config, error) {
	_ = godotenv.Load()
	cfg := Config{
		MailProvider: getEnv("MAIL_PROVIDER", "console"),
		SMTPHost:     os.Getenv("SMTP_HOST"),
		SMTPPort:     getEnv("SMTP_PORT", "465"),
		SMTPUser:     os.Getenv("SMTP_USER"),
		SMTPPass:     os.Getenv("SMTP_PASS"),
		SMTPFrom:     os.Getenv("SMTP_FROM"),

		TencentSESSecretID:   os.Getenv("TENCENT_SES_SECRET_ID"),
		TencentSESSecretKey:  os.Getenv("TENCENT_SES_SECRET_KEY"),
		TencentSESRegion:     os.Getenv("TENCENT_SES_REGION"),
		TencentSESFrom:       os.Getenv("TENCENT_SES_FROM"),
		TencentSESTemplateID: getEnvUint("TENCENT_SES_TEMPLATE_ID"),

		Port:         getEnv("PORT", "8090"),
		DBDSN:        getEnv("DB_DSN", "root:modernchat_sso_dev@tcp(127.0.0.1:3307)/modernchat_sso?charset=utf8mb4&parseTime=true&loc=Local"),
		JWTSecret:    os.Getenv("JWT_SECRET"),
		CookieSecure: getEnv("COOKIE_SECURE", "false") == "true",

		WebUpdate:         getEnv("WEB_UPDATE", "true") == "true",
		ReleaseAPIURL:     os.Getenv("RELEASE_API_URL"),
		ReleaseDLURL:      os.Getenv("RELEASE_DL_URL"),
		ReleaseRepoGitHub: os.Getenv("RELEASE_REPO_GITHUB"),
		GitHubToken:       os.Getenv("GITHUB_TOKEN"),
	}
	// 不提供不安全的默认值：宁可起不来，也不能用空/弱密钥签 token。
	// ≥32 字节与主仓库生产要求一致（运维手册 §JWT_SECRET）
	if len(cfg.JWTSecret) < 32 {
		return cfg, fmt.Errorf("JWT_SECRET 未设置或不足 32 字节（生成：openssl rand -hex 32）")
	}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvUint(key string) uint64 {
	n, _ := strconv.ParseUint(os.Getenv(key), 10, 64)
	return n
}
