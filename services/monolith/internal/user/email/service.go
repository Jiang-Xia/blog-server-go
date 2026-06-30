// Package email 邮箱验证码发送与校验（Redis + SMTP）。
package email

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"math/big"
	"net"
	"net/smtp"
	"strings"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/pkg/redisutil"
)

const (
	codeTTLSec       = 5 * 60
	frequencyTTLSec  = 60
)

// Service 邮件验证码服务。
type Service struct {
	cfg   *config.Config
	redis *redisutil.Store
}

// NewService 构造 EmailService。
func NewService(cfg *config.Config, redis *redisutil.Store) *Service {
	return &Service{cfg: cfg, redis: redis}
}

// SendCode 生成 6 位验证码写入 Redis 并通过 SMTP 发送；未配置邮件时返回 errcode。
func (s *Service) SendCode(ctx context.Context, emailAddr, codeType string) error {
	if !s.cfg.Mail.MailConfigured() {
		return errcode.WithMessage(errcode.InternalError, "邮件发送未配置")
	}
	code, err := generateVerificationCode()
	if err != nil {
		return err
	}
	key := verificationKey(codeType, emailAddr)
	if err := s.redis.Set(ctx, key, code, codeTTLSec); err != nil {
		return err
	}
	subject, body := mailContent(codeType, code)
	return s.sendSMTP(emailAddr, subject, body)
}

// VerifyCode 校验验证码，成功后删除 Redis 键（单次有效）。
func (s *Service) VerifyCode(ctx context.Context, emailAddr, code, codeType string) error {
	key := verificationKey(codeType, emailAddr)
	stored, err := s.redis.Get(ctx, key)
	if err != nil {
		return err
	}
	if stored == "" {
		return errcode.WithMessage(errcode.InvalidParam, "验证码已过期或不存在")
	}
	if stored != code {
		return errcode.WithMessage(errcode.InvalidParam, "验证码错误")
	}
	return s.redis.Del(ctx, key)
}

// CheckSendFrequency 60 秒内同邮箱同类型不可重复发送。
func (s *Service) CheckSendFrequency(ctx context.Context, emailAddr, codeType string) error {
	key := frequencyKey(codeType, emailAddr)
	last, err := s.redis.Get(ctx, key)
	if err != nil {
		return err
	}
	if last != "" {
		return errcode.WithMessage(errcode.InvalidParam, "请求过于频繁，请稍后再试")
	}
	return s.redis.Set(ctx, key, "1", frequencyTTLSec)
}

func verificationKey(codeType, emailAddr string) string {
	return fmt.Sprintf("email_verification_code:%s:%s", codeType, emailAddr)
}

func frequencyKey(codeType, emailAddr string) string {
	return fmt.Sprintf("email_send_frequency:%s:%s", codeType, emailAddr)
}

func generateVerificationCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()+100000), nil
}

func (s *Service) sendSMTP(to, subject, htmlBody string) error {
	m := s.cfg.Mail
	host := m.Host
	if host == "" {
		host = "smtp.163.com"
	}
	port := m.Port
	if port == 0 {
		port = 465
	}
	addr := fmt.Sprintf("%s:%d", host, port)
	from := m.User
	msg := buildMIME(from, to, subject, htmlBody)
	auth := smtp.PlainAuth("", m.User, m.Pass, host)
	// 465 端口需 TLS；此处使用 smtp.SendMail（STARTTLS 或明文视端口而定，与 Nest nodemailer secure 对齐时优先 465+TLS）。
	if port == 465 {
		return sendMailTLS(addr, auth, from, []string{to}, msg)
	}
	return smtp.SendMail(addr, auth, from, []string{to}, msg)
}

func buildMIME(from, to, subject, htmlBody string) []byte {
	headers := []string{
		fmt.Sprintf("From: \"江夏的博客系统\" <%s>", from),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
		"",
		htmlBody,
	}
	return []byte(strings.Join(headers, "\r\n"))
}

func mailContent(codeType, code string) (subject, html string) {
	wrap := func(title, hint, color string) string {
		return fmt.Sprintf(`<div style="max-width:600px;margin:0 auto;padding:20px;font-family:Arial,sans-serif;">
<h2 style="color:#333;">%s</h2>
<p>%s</p>
<div style="background:#f5f5f5;padding:20px;text-align:center;margin:20px 0;">
<span style="font-size:24px;color:%s;font-weight:bold;letter-spacing:3px;">%s</span>
</div>
<p style="color:#666;">验证码有效期为5分钟，请尽快使用。</p>
</div>`, title, hint, color, code)
	}
	switch codeType {
	case "login":
		return "登录验证 - 邮箱验证码", wrap("登录验证", "您正在进行邮箱登录，验证码为：", "#28a745")
	case "reset":
		return "密码重置 - 邮箱验证码", wrap("密码重置验证", "您正在进行密码重置，验证码为：", "#dc3545")
	default:
		return "欢迎注册 - 邮箱验证码", wrap("欢迎注册我们的服务", "您正在进行邮箱注册，验证码为：", "#007bff")
	}
}

// sendMailTLS 通过 SMTPS（465）发送邮件。
func sendMailTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: host})
	if err != nil {
		return err
	}
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Close()
	if auth != nil {
		if err = client.Auth(auth); err != nil {
			return err
		}
	}
	if err = client.Mail(from); err != nil {
		return err
	}
	for _, rcpt := range to {
		if err = client.Rcpt(rcpt); err != nil {
			return err
		}
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err = w.Write(msg); err != nil {
		return err
	}
	if err = w.Close(); err != nil {
		return err
	}
	return client.Quit()
}
