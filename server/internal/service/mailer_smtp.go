package service

import (
	"cmp"
	"context"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"mime"
	"net"
	"net/smtp"
	"strings"
	"time"
)

const (
	mailTimeout  = 10 * time.Second
	mailFromName = "Modernchat SSO"
)

// SMTPMailer 标准 SMTP 通道：465 隐式 TLS；587/25 明文连接后 STARTTLS 升级再认证。
// 明文连接上拒绝带凭据发送（与主仓库 mailer.go 同语义）。
type SMTPMailer struct {
	host, port, user, pass, from string
}

// NewSMTPMailer 构造；端口缺省 465。
func NewSMTPMailer(host, port, user, pass, from string) *SMTPMailer {
	return &SMTPMailer{host: host, port: cmp.Or(port, "465"), user: user, pass: pass, from: from}
}

// Enabled 工厂仅在配置完整时构造 SMTPMailer，此处恒为 true。
func (m *SMTPMailer) Enabled() bool { return true }

// Send 发送纯文本邮件。
// net/smtp 不感知 ctx：只在拨号前检查取消；其余超时由 dialer/SetDeadline（mailTimeout）兜底。
func (m *SMTPMailer) Send(ctx context.Context, msg MailMessage) error {
	if m.host == "" || m.from == "" {
		return errors.New("SMTP 未配置完整（SMTP_HOST/SMTP_FROM）")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return sendMail(m.host, m.port, m.user, m.pass, m.from, msg.To,
		buildPlainMessage(mailFromName, m.from, msg.To, msg.Subject, msg.Text))
}

// buildPlainMessage 通用纯文本邮件组装（RFC 2045：base64 正文每行 ≤76 字符）。
func buildPlainMessage(fromName, from, to, subject, body string) []byte {
	var b strings.Builder
	b.WriteString("From: " + fromName + " <" + from + ">\r\n")
	b.WriteString("To: <" + to + ">\r\n")
	b.WriteString("Subject: " + mime.BEncoding.Encode("UTF-8", subject) + "\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	b.WriteString("Content-Transfer-Encoding: base64\r\n")
	b.WriteString("\r\n")

	enc := base64.StdEncoding.EncodeToString([]byte(body))
	for len(enc) > 76 {
		b.WriteString(enc[:76] + "\r\n")
		enc = enc[76:]
	}
	b.WriteString(enc + "\r\n")
	return []byte(b.String())
}

func sendMail(host, port, user, pass, from, to string, msg []byte) error {
	addr := net.JoinHostPort(host, port)
	tlsCfg := &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}
	dialer := &net.Dialer{Timeout: mailTimeout}

	var conn net.Conn
	var err error
	if port == "465" {
		conn, err = tls.DialWithDialer(dialer, "tcp", addr, tlsCfg)
	} else {
		conn, err = dialer.Dial("tcp", addr)
	}
	if err != nil {
		return fmt.Errorf("dial %s: %w", addr, err)
	}
	_ = conn.SetDeadline(time.Now().Add(mailTimeout))

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("smtp client: %w", err)
	}
	defer func() { _ = c.Close() }()

	if port != "465" {
		if ok, _ := c.Extension("STARTTLS"); ok {
			if err := c.StartTLS(tlsCfg); err != nil {
				return fmt.Errorf("starttls: %w", err)
			}
		} else if user != "" {
			// 明文连接上不允许发送凭据（smtp.PlainAuth 也会拒绝），直接报错更清晰
			return errors.New("server does not support STARTTLS; use port 465 for implicit TLS")
		}
	}

	if user != "" {
		auth := smtp.PlainAuth("", user, pass, host)
		if err := c.Auth(auth); err != nil {
			return fmt.Errorf("auth: %w", err)
		}
	}
	if err := c.Mail(from); err != nil {
		return fmt.Errorf("mail from: %w", err)
	}
	if err := c.Rcpt(to); err != nil {
		return fmt.Errorf("rcpt to: %w", err)
	}
	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("data: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("write body: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("close body: %w", err)
	}
	return c.Quit()
}
