package email

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"html/template"
	"time"

	"github.com/luyb177/silent-sign-backend/common/mail"
)

// EmailSender 邮件发送接口
type EmailSender interface {
	// SendVerifyCode 发送验证码邮件
	SendVerifyCode(ctx context.Context, to, code string, expireMinutes int) error

	// SendWelcomeEmail 发送欢迎邮件
	SendWelcomeEmail(ctx context.Context, to, username string) error

	// SendLoginNotification 发送登录通知邮件
	SendLoginNotification(ctx context.Context, to, username, ip, location string) error
}

type DefaultEmailSender struct {
	mailer *mail.Mailer
}

func NewEmailSender(m *mail.Mailer) EmailSender {
	return &DefaultEmailSender{
		mailer: m,
	}
}

func (s *DefaultEmailSender) SendVerifyCode(ctx context.Context, to, code string, expireMinutes int) error {
	subject := "【Silent Sign】邮箱验证码"
	data := map[string]interface{}{
		"Code":          code,
		"ExpireMinutes": expireMinutes,
		"Now":           time.Now().Format(time.RFC3339),
	}

	return s.renderAndSend(ctx, subject, verifyCodeTmpl, data, []string{to})
}

func (s *DefaultEmailSender) SendWelcomeEmail(ctx context.Context, to, username string) error {
	subject := "【Silent Sign】欢迎加入"
	data := map[string]interface{}{
		"Username": username,
		"Now":      time.Now().Format(time.RFC3339),
	}

	return s.renderAndSend(ctx, subject, welcomeTmpl, data, []string{to})
}

func (s *DefaultEmailSender) SendLoginNotification(ctx context.Context, to, username, ip, location string) error {
	subject := "【Silent Sign】登录通知"
	data := map[string]interface{}{
		"Username": username,
		"IP":       ip,
		"Location": location,
		"Now":      time.Now().Format(time.RFC3339),
	}

	return s.renderAndSend(ctx, subject, loginNotificationTmpl, data, []string{to})
}

// 内部渲染并发送
func (s *DefaultEmailSender) renderAndSend(ctx context.Context, subject string, tmpl *template.Template, data interface{}, to []string) error {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("execute email template failed: %w", err)
	}
	return s.mailer.Send(subject, buf.String(), to)
}
