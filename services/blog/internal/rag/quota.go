package rag

import (
	"context"
	"fmt"
	"time"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/pkg/redisutil"
)

// QuotaUsage 今日配额使用情况。
type QuotaUsage struct {
	Used      int `json:"used"`
	Limit     int `json:"limit"`
	Remaining int `json:"remaining"`
}

// QuotaService Redis 日配额。
type QuotaService struct {
	cfg   *config.Config
	redis *redisutil.Store
}

// NewQuotaService 构造 QuotaService。
func NewQuotaService(cfg *config.Config, redis *redisutil.Store) *QuotaService {
	return &QuotaService{cfg: cfg, redis: redis}
}

func (s *QuotaService) dailyKey(uid int) string {
	return fmt.Sprintf("%s:%d:%s", QueryRedisPrefix, uid, time.Now().Format("2006-01-02"))
}

func (s *QuotaService) dailyLimit() int {
	return s.cfg.Rag.RagDailyQuotaOrDefault()
}

// GetUsage 返回今日配额使用情况。
func (s *QuotaService) GetUsage(ctx context.Context, uid int) (QuotaUsage, error) {
	limit := s.dailyLimit()
	raw, err := s.redis.Get(ctx, s.dailyKey(uid))
	if err != nil {
		return QuotaUsage{}, err
	}
	used := int(redisutil.ParseInt(raw))
	remaining := limit - used
	if remaining < 0 {
		remaining = 0
	}
	return QuotaUsage{Used: used, Limit: limit, Remaining: remaining}, nil
}

// AssertQuota 检查配额，不足则返回 429 业务错误。
func (s *QuotaService) AssertQuota(ctx context.Context, uid int) error {
	usage, err := s.GetUsage(ctx, uid)
	if err != nil {
		return err
	}
	if usage.Used >= usage.Limit {
		return errcode.WithMessage(errcode.TooManyRequests,
			"今日 AI 助手问答次数已达上限（%d 次）", usage.Limit)
	}
	return nil
}

// Consume LLM 流成功启动后扣次。
func (s *QuotaService) Consume(ctx context.Context, uid int) error {
	key := s.dailyKey(uid)
	count, err := s.redis.Incr(ctx, key)
	if err != nil {
		return err
	}
	if count == 1 {
		now := time.Now()
		end := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
		ttl := int(end.Sub(now).Seconds()) + 60
		if ttl < 60 {
			ttl = 60
		}
		return s.redis.Expire(ctx, key, ttl)
	}
	return nil
}
