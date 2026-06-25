package infrastructure

import (
	"context"
	"fmt"
	"net/smtp"

	"github.com/ZY0506/gin-scaffold/config"
	bizErrors "github.com/ZY0506/gin-scaffold/internal/pkg/errors"
)

type SMTPSender struct {
	cfg *config.EmailConfig
}

func NewSMTPSender(cfg *config.EmailConfig) *SMTPSender {
	return &SMTPSender{cfg: cfg}
}

func (s *SMTPSender) Send(ctx context.Context, to, subject, body string) error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)

	// 认证信息
	auth := smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)

	// 构建邮件内容
	msg := []byte(fmt.Sprintf("From: %s <%s>\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: text/plain; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n", s.cfg.FromName, s.cfg.Username, to, subject, body))

	// 发送
	err := smtp.SendMail(addr, auth, s.cfg.Username, []string{to}, msg)
	if err != nil {
		return bizErrors.Wrap(err, bizErrors.ErrInternal, "发送邮件失败")
	}

	return nil
}
