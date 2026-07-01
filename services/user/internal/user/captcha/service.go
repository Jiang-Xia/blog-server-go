// Package captcha 图形验证码生成与 Redis 校验。
package captcha

import (
	"context"
	"encoding/base64"
	"strings"

	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/pkg/redisutil"
	svgcaptcha "github.com/reu98/go-svg-captcha"
	"github.com/google/uuid"
)

const (
	rateWindowSec    = 60
	rateMaxPerWindow = 5
	verifyMaxAttempt = 5
	captchaTTL       = 120
)

// Service 验证码服务。
type Service struct {
	redis *redisutil.Store
}

// NewService 构造 Service。
func NewService(redis *redisutil.Store) *Service {
	return &Service{redis: redis}
}

// CaptchaResult 生成结果，字段与 Nest /user/authCode 响应对齐。
type CaptchaResult struct {
	ID            string
	CaptchaBase64 string // SVG 内容的 base64 编码
}

// Create 生成验证码并写入 Redis。
func (s *Service) Create(ctx context.Context) (*CaptchaResult, error) {
	result, err := svgcaptcha.CreateByText(svgcaptcha.OptionText{
		Size:             4,
		Width:            100,
		Height:           48,
		IsColor:          true,
		Curve:            2,
		IgnoreCharacters: "0o1iIlL",
		CharactersPreset: "ABCDEFGHJKLMNPQRSTUVWXYZ23456789",
	})
	if err != nil {
		return nil, err
	}

	id := strings.ReplaceAll(uuid.NewString(), "-", "")
	if err := s.redis.Set(ctx, captchaKey(id), strings.ToLower(result.Text), captchaTTL); err != nil {
		return nil, err
	}
	return &CaptchaResult{
		ID:            id,
		CaptchaBase64: base64.StdEncoding.EncodeToString([]byte(result.Data)),
	}, nil
}

// AssertRateLimit 生成频率限制。
func (s *Service) AssertRateLimit(ctx context.Context, identity string) error {
	key := "captcha:rate:" + redisSafe(identity)
	n, err := s.redis.Incr(ctx, key)
	if err != nil {
		return err
	}
	if n == 1 {
		_ = s.redis.Expire(ctx, key, rateWindowSec)
	}
	if n > rateMaxPerWindow {
		return errcode.WithMessage(errcode.CaptchaRefresh, "验证码获取过于频繁，请稍后再试")
	}
	return nil
}

// Verify 校验验证码。
func (s *Service) Verify(ctx context.Context, id, answer string) error {
	if id == "" {
		return errcode.WithMessage(errcode.CaptchaRefresh, "验证码已过期")
	}
	stored, err := s.redis.Get(ctx, captchaKey(id))
	if err != nil {
		return err
	}
	if stored == "" {
		return errcode.WithMessage(errcode.CaptchaRefresh, "验证码已过期")
	}
	if strings.EqualFold(strings.TrimSpace(answer), stored) {
		_ = s.redis.Del(ctx, captchaKey(id), attemptKey(id))
		return nil
	}
	ak := attemptKey(id)
	count, _ := s.redis.Incr(ctx, ak)
	if count == 1 {
		_ = s.redis.Expire(ctx, ak, captchaTTL)
	}
	if count >= verifyMaxAttempt {
		_ = s.redis.Del(ctx, captchaKey(id), ak)
		return errcode.WithMessage(errcode.InvalidParam, "验证码错误次数过多，请刷新后重试")
	}
	return errcode.WithMessage(errcode.InvalidParam, "验证码错误")
}

func captchaKey(id string) string { return "captcha:" + id }
func attemptKey(id string) string { return "captcha:attempt:" + id }

func redisSafe(v string) string {
	if v == "" {
		return "unknown"
	}
	var b strings.Builder
	for _, r := range v {
		if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '.' || r == '_' || r == '-' {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	return b.String()
}
