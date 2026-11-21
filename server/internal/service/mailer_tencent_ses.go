package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	ses "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ses/v20201002"
)

// TencentSESMailer 腾讯云 SES（邮件推送）通道。
// SES 是模板发送制（自定义正文需特殊申请）：在 SES 控制台建一个正文为 {{content}} 的模板
// （审核通过后），TemplateID 填入 TENCENT_SES_TEMPLATE_ID；Subject 直接透传，
// 邮件正文经模板变量 {{content}} 注入；TriggerType=1（验证码等触发类即时邮件）。
type TencentSESMailer struct {
	client     *ses.Client
	from       string
	templateID uint64
}

// TencentSESConfig 腾讯云 SES 接入参数（全部来自环境变量，凭据严禁入库）。
type TencentSESConfig struct {
	SecretID   string
	SecretKey  string
	Region     string // 缺省 ap-guangzhou
	From       string // 发件地址（可带别名："别名 <addr>"，须已在 SES 控制台验证）
	TemplateID uint64 // 正文为 {{content}} 的已过审模板
}

// NewTencentSESMailer 构造；参数不完整返回错误（由工厂决定退回 console）。
func NewTencentSESMailer(cfg TencentSESConfig) (*TencentSESMailer, error) {
	if cfg.SecretID == "" || cfg.SecretKey == "" || cfg.From == "" || cfg.TemplateID == 0 {
		return nil, errors.New("腾讯云 SES 配置不完整（TENCENT_SES_SECRET_ID/SECRET_KEY/FROM/TEMPLATE_ID）")
	}
	if cfg.Region == "" {
		cfg.Region = "ap-guangzhou"
	}
	client, err := ses.NewClient(common.NewCredential(cfg.SecretID, cfg.SecretKey), cfg.Region, profile.NewClientProfile())
	if err != nil {
		return nil, fmt.Errorf("ses client: %w", err)
	}
	return &TencentSESMailer{client: client, from: cfg.From, templateID: cfg.TemplateID}, nil
}

// Enabled 工厂仅在配置完整时构造，此处恒为 true。
func (m *TencentSESMailer) Enabled() bool { return true }

// Send 经模板发送：正文整体作为 {{content}} 变量注入。
func (m *TencentSESMailer) Send(ctx context.Context, msg MailMessage) error {
	data, err := json.Marshal(map[string]string{"content": msg.Text})
	if err != nil {
		return fmt.Errorf("template data: %w", err)
	}
	req := ses.NewSendEmailRequest()
	req.FromEmailAddress = common.StringPtr(m.from)
	req.Destination = []*string{common.StringPtr(msg.To)}
	req.Subject = common.StringPtr(msg.Subject)
	req.TriggerType = common.Uint64Ptr(1) // 触发类：验证码等即时邮件
	req.Template = &ses.Template{
		TemplateID:   common.Uint64Ptr(m.templateID),
		TemplateData: common.StringPtr(string(data)),
	}
	if _, err := m.client.SendEmailWithContext(ctx, req); err != nil {
		return fmt.Errorf("ses send: %w", err)
	}
	return nil
}
