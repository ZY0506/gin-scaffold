package domain

import "context"

// EmailSender 邮件发送接口
type EmailSender interface {
	Send(ctx context.Context, to, subject, body string) error
}

// CodeStore 验证码存储接口
type CodeStore interface {
	// Set 存储验证码，ttl 为过期时间
	Set(ctx context.Context, email, code string, ttl int) error
	// Get 获取验证码，返回空字符串表示不存在或已过期
	Get(ctx context.Context, email string) (string, error)
	// Del 删除验证码（验证成功后清除）
	Del(ctx context.Context, email string) error
}
