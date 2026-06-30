// Package redisutil 封装 rueidis 常用命令，供 auth/captcha 等模块使用。
package redisutil

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/rueidis"
)

// Store Redis 字符串/计数操作封装。
type Store struct {
	client rueidis.Client
}

// New 构造 Store。
func New(client rueidis.Client) *Store {
	return &Store{client: client}
}

// Get 读取字符串键，不存在返回空串。
func (s *Store) Get(ctx context.Context, key string) (string, error) {
	resp := s.client.Do(ctx, s.client.B().Get().Key(key).Build())
	if rueidis.IsRedisNil(resp.Error()) {
		return "", nil
	}
	return resp.ToString()
}

// Set 写入字符串键并设置 TTL（秒）。
func (s *Store) Set(ctx context.Context, key, value string, ttlSec int) error {
	return s.client.Do(ctx, s.client.B().Set().Key(key).Value(value).ExSeconds(int64(ttlSec)).Build()).Error()
}

// SetNX 仅在键不存在时写入。
func (s *Store) SetNX(ctx context.Context, key, value string, ttlSec int) (bool, error) {
	resp := s.client.Do(ctx, s.client.B().Set().Key(key).Value(value).Nx().ExSeconds(int64(ttlSec)).Build())
	if resp.Error() != nil {
		return false, resp.Error()
	}
	val, err := resp.ToString()
	if err != nil {
		return false, err
	}
	return val == "OK", nil
}

// Del 删除键。
func (s *Store) Del(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return s.client.Do(ctx, s.client.B().Del().Key(keys...).Build()).Error()
}

// Incr 自增并返回新值。
func (s *Store) Incr(ctx context.Context, key string) (int64, error) {
	return s.client.Do(ctx, s.client.B().Incr().Key(key).Build()).AsInt64()
}

// Expire 设置过期时间（秒）。
func (s *Store) Expire(ctx context.Context, key string, ttlSec int) error {
	return s.client.Do(ctx, s.client.B().Expire().Key(key).Seconds(int64(ttlSec)).Build()).Error()
}

// TTL 返回键剩余秒数。
func (s *Store) TTL(ctx context.Context, key string) (time.Duration, error) {
	sec, err := s.client.Do(ctx, s.client.B().Ttl().Key(key).Build()).AsInt64()
	if err != nil {
		return 0, err
	}
	return time.Duration(sec) * time.Second, nil
}

// ParseInt 解析 Redis 计数字符串。
func ParseInt(v string) int64 {
	n, _ := strconv.ParseInt(v, 10, 64)
	return n
}
