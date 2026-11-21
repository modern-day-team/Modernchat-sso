// Package config 配置加载：.env 可选，环境变量优先。
package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config 服务运行配置。
type Config struct {
	Port         string // HTTP 监听端口
	DBDSN        string // MySQL 连接串
	JWTSecret    string // JWT 签名密钥
	CookieSecure bool   // mc_auth cookie 是否带 Secure（生产必须 true）
}

// Load 读取配置；.env 不存在时静默跳过（生产用环境变量注入）。
func Load() (Config, error) {
	_ = godotenv.Load()
	cfg := Config{
		Port:         getEnv("PORT", "8090"),
		DBDSN:        getEnv("DB_DSN", "root:modernchat_sso_dev@tcp(127.0.0.1:3307)/modernchat_sso?charset=utf8mb4&parseTime=true&loc=Local"),
		JWTSecret:    os.Getenv("JWT_SECRET"),
		CookieSecure: getEnv("COOKIE_SECURE", "false") == "true",
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
