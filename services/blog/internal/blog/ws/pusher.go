package ws

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/rueidis"
)

// Pusher 实时推送接口，供 notification 等模块注入。
type Pusher interface {
	PushToUser(ctx context.Context, uid uint64, msgType string, seq uint64, data interface{}) error
}

// RealtimePusher 本进程 Hub 直推 + Redis pub/sub 跨模块推送。
type RealtimePusher struct {
	hub *Hub
	rds rueidis.Client
}

// NewRealtimePusher 构造 RealtimePusher。
func NewRealtimePusher(hub *Hub, rds rueidis.Client) *RealtimePusher {
	return &RealtimePusher{hub: hub, rds: rds}
}

// PushToUser 组装 WS 消息并推送给指定用户（本进程 Hub）。
func (p *RealtimePusher) PushToUser(_ context.Context, uid uint64, msgType string, seq uint64, data interface{}) error {
	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}
	payload, err := json.Marshal(Message{Type: msgType, Seq: seq, Data: raw})
	if err != nil {
		return err
	}
	p.hub.PublishToUser(uid, payload)
	return nil
}

// PublishRedis 经 Redis pub/sub 推送（跨模块/跨实例）。
func (p *RealtimePusher) PublishRedis(ctx context.Context, uid uint64, topic string, body []byte) error {
	env, err := json.Marshal(RedisPushMsg{UserID: uid, Topic: topic, Body: body})
	if err != nil {
		return err
	}
	return p.rds.Do(ctx, p.rds.B().Publish().Channel(ChannelRealtimePush).Message(string(env)).Build()).Error()
}

// StartRedisSubscriber 订阅 realtime:push 并转发到 Hub；ctx 取消时退出。
func StartRedisSubscriber(ctx context.Context, client rueidis.Client, hub *Hub) {
	go func() {
		for ctx.Err() == nil {
			err := client.Dedicated(func(c rueidis.DedicatedClient) error {
				errCh := c.SetPubSubHooks(rueidis.PubSubHooks{
					OnMessage: func(m rueidis.PubSubMessage) {
						hub.DispatchRaw([]byte(m.Message))
					},
				})
				if err := c.Do(ctx, c.B().Subscribe().Channel(ChannelRealtimePush).Build()).Error(); err != nil {
					return err
				}
				select {
				case <-ctx.Done():
					return nil
				case err := <-errCh:
					return err
				}
			})
			if ctx.Err() != nil {
				return
			}
			if err != nil {
				time.Sleep(2 * time.Second)
			}
		}
	}()
}
