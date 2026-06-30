// Package captcha 图形验证码生成与 Redis 校验。
package captcha

import (
	"context"
	"strings"

	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/pkg/redisutil"
	"github.com/mojocn/base64Captcha"
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

// CaptchaResult 生成结果。
type CaptchaResult struct {
	ID  string
	SVG string // 实际为 base64 图片，字段名与 Nest authCode 响应对齐
}

type redisCaptchaStore struct {
	redis *redisutil.Store
	ctx   context.Context
}

func (s *redisCaptchaStore) Set(id string, value string) error {
	return s.redis.Set(s.ctx, captchaKey(id), strings.ToLower(value), captchaTTL)
}

func (s *redisCaptchaStore) Get(id string, clear bool) string {
	v, _ := s.redis.Get(s.ctx, captchaKey(id))
	if clear && v != "" {
		_ = s.redis.Del(s.ctx, captchaKey(id), attemptKey(id))
	}
	return v
}

func (s *redisCaptchaStore) Verify(id, answer string, clear bool) bool {
	stored := s.Get(id, false)
	if stored == "" {
		return false
	}
	ok := strings.EqualFold(strings.TrimSpace(answer), stored)
	if ok && clear {
		_ = s.redis.Del(s.ctx, captchaKey(id), attemptKey(id))
	}
	return ok
}

// Create 生成验证码并写入 Redis。
func (s *Service) Create(ctx context.Context) (*CaptchaResult, error) {
	driver := base64Captcha.NewDriverString(
		48, 100, 2, 0,
		4, "ABCDEFGHJKLMNPQRSTUVWXYZ23456789", nil, nil, nil,
	)
	store := &redisCaptchaStore{redis: s.redis, ctx: ctx}
	c := base64Captcha.NewCaptcha(driver, store)
	id, b64, _, err := c.Generate()
	if err != nil {
		return nil, err
	}
	return &CaptchaResult{ID: id, SVG: b64}, nil
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
