// Package wspush 经 Redis pub/sub 向 blog-service WS Hub 推送（跨服务）。
package wspush

import (
	"context"
	"encoding/json"

	"github.com/redis/rueidis"
)

const channelRealtimePush = "realtime:push"

// Pusher 实时推送接口，与 blog/ws.Pusher 对齐。
type Pusher interface {
	PushToUser(ctx context.Context, uid uint64, msgType string, seq uint64, data interface{}) error
}

// RedisPusher 无本地 Hub 时经 Redis 转发到 blog WS。
type RedisPusher struct {
	rds rueidis.Client
}

// NewRedisPusher 构造 RedisPusher。
func NewRedisPusher(rds rueidis.Client) *RedisPusher {
	return &RedisPusher{rds: rds}
}

type wsMessage struct {
	Type string          `json:"type"`
	Seq  uint64          `json:"seq,omitempty"`
	Data json.RawMessage `json:"data,omitempty"`
}

type redisPushMsg struct {
	UserID uint64          `json:"userId"`
	Topic  string          `json:"topic,omitempty"`
	Body   json.RawMessage `json:"body"`
}

// PushToUser 组装 WS 消息并 publish 到 realtime:push。
func (p *RedisPusher) PushToUser(ctx context.Context, uid uint64, msgType string, seq uint64, data interface{}) error {
	if p.rds == nil {
		return nil
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}
	payload, err := json.Marshal(wsMessage{Type: msgType, Seq: seq, Data: raw})
	if err != nil {
		return err
	}
	env, err := json.Marshal(redisPushMsg{UserID: uid, Body: payload})
	if err != nil {
		return err
	}
	return p.rds.Do(ctx, p.rds.B().Publish().Channel(channelRealtimePush).Message(string(env)).Build()).Error()
}
