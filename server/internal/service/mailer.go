// 邮件发送层：SSO 的邮箱验证会接入不同厂商。
// 业务只依赖 Mailer 接口；新增厂商 = 加一个实现 + buildProvider 加一个分支，业务代码零改动。
// 通道：console（仅本地开发，内容打日志）| smtp（465 隐式 TLS / 587 STARTTLS）| tencent_ses（腾讯云 SES 模板）| 其他厂商按同接口扩展。
// MAIL_PROVIDER 支持逗号分隔多通道，按序故障转移：如 tencent_ses,smtp —— 主选 SES 失败自动降级 SMTP。
package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/modern-day-team/Modernchat-sso/server/internal/config"
)

// MailMessage 一封纯文本邮件（SSO 的验证码/重置链接都是纯文本；HTML 需要时再扩展）。
type MailMessage struct {
	To      string
	Subject string
	Text    string
}

// Mailer 邮件发送抽象。
type Mailer interface {
	Send(ctx context.Context, msg MailMessage) error
}

// NewMailer 按配置装配通道链：MAIL_PROVIDER 逗号分隔，按序故障转移。
// 配置不完整的通道跳过并告警；全部不可用退回 console（error 级日志，生产误配必须显眼）。
func NewMailer(cfg config.Config) Mailer {
	var chain []Mailer
	var skipped []string
	for _, name := range strings.Split(cfg.MailProvider, ",") {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		m, err := buildProvider(name, cfg)
		if err != nil {
			skipped = append(skipped, fmt.Sprintf("%s: %v", name, err))
			continue
		}
		chain = append(chain, &namedMailer{name: name, inner: m})
	}
	if len(chain) == 0 {
		slog.Error("无可用邮件通道，退回 console（不发真信）", "MAIL_PROVIDER", cfg.MailProvider, "skipped", skipped)
		return NewConsoleMailer()
	}
	if len(skipped) > 0 {
		slog.Warn("部分邮件通道配置不完整已跳过", "skipped", skipped)
	}
	if len(chain) == 1 {
		return chain[0]
	}
	return &failoverMailer{chain: chain}
}

// buildProvider 单通道构造；配置不完整返回错误，由调用方决定跳过或告警。
func buildProvider(name string, cfg config.Config) (Mailer, error) {
	switch name {
	case "console":
		return NewConsoleMailer(), nil
	case "smtp":
		if cfg.SMTPHost == "" {
			return nil, errors.New("SMTP_HOST 为空")
		}
		return NewSMTPMailer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPFrom), nil
	case "tencent_ses":
		return NewTencentSESMailer(TencentSESConfig{
			SecretID:   cfg.TencentSESSecretID,
			SecretKey:  cfg.TencentSESSecretKey,
			Region:     cfg.TencentSESRegion,
			From:       cfg.TencentSESFrom,
			TemplateID: cfg.TencentSESTemplateID,
		})
	default:
		return nil, fmt.Errorf("未知通道（可用：console/smtp/tencent_ses）")
	}
}

// namedMailer 带通道名的包装：降级日志与聚合错误里能定位到具体厂商。
type namedMailer struct {
	name  string
	inner Mailer
}

func (m *namedMailer) Send(ctx context.Context, msg MailMessage) error {
	if err := m.inner.Send(ctx, msg); err != nil {
		return fmt.Errorf("%s: %w", m.name, err)
	}
	return nil
}

// failoverMailer 多通道按序故障转移：主选失败自动降级下一通道，全部失败才报错（错误聚合）。
type failoverMailer struct {
	chain []Mailer
}

func (m *failoverMailer) Send(ctx context.Context, msg MailMessage) error {
	var errs []error
	for _, p := range m.chain {
		if err := p.Send(ctx, msg); err == nil {
			return nil
		} else {
			slog.Warn("邮件通道发送失败，尝试下一通道", "err", err, "to", msg.To)
			errs = append(errs, err)
		}
	}
	return fmt.Errorf("所有邮件通道均失败: %w", errors.Join(errs...))
}
