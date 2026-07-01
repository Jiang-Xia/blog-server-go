// consumer Redis Stream 消费者循环，分发至已注册 handler。
package event

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/redis/rueidis"
	"go.uber.org/zap"
)

// Handler 单类事件消费函数；返回 error 时不 ACK（留 pending）。
type Handler func(ctx context.Context, payload json.RawMessage) error

// Consumer Redis Stream 消费器，支持注册多 handler。
type Consumer struct {
	rds      rueidis.Client
	log      *zap.Logger
	group    string
	consumer string
	handlers map[string]Handler
}

// NewConsumer 构造 Consumer。
func NewConsumer(rds rueidis.Client, log *zap.Logger, group string) *Consumer {
	hostname, _ := os.Hostname()
	return &Consumer{
		rds:      rds,
		log:      log,
		group:    group,
		consumer: fmt.Sprintf("%s-%d", hostname, os.Getpid()),
		handlers: make(map[string]Handler),
	}
}

// Register 注册事件 handler。
func (c *Consumer) Register(eventType string, h Handler) {
	c.handlers[eventType] = h
}

// Start 启动消费循环（独立 goroutine，ctx 取消时退出）。
func (c *Consumer) Start(ctx context.Context) {
	go func() {
		for ctx.Err() == nil {
			err := c.rds.Dedicated(func(d rueidis.DedicatedClient) error {
				if err := c.ensureGroup(ctx); err != nil {
					return err
				}
				for ctx.Err() == nil {
					resp := d.Do(ctx, d.B().Xreadgroup().
						Group(c.group, c.consumer).
						Count(10).Block(5000).
						Streams().Key(StreamBlogEvents).Id(">").
						Build())
					if err := resp.Error(); err != nil {
						if rueidis.IsRedisNil(err) {
							continue
						}
						return err
					}
					entries, err := parseXReadGroup(resp)
					if err != nil {
						return err
					}
					for _, entry := range entries {
						c.processEntry(ctx, d, entry)
					}
				}
				return nil
			})
			if ctx.Err() != nil {
				return
			}
			if err != nil {
				c.log.Warn("event consumer disconnected", zap.Error(err))
				time.Sleep(2 * time.Second)
			}
		}
	}()
}

func (c *Consumer) ensureGroup(ctx context.Context) error {
	err := c.rds.Do(ctx, c.rds.B().XgroupCreate().Key(StreamBlogEvents).Group(c.group).Id("$").Mkstream().Build()).Error()
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), "BUSYGROUP") {
		return nil
	}
	return err
}

func (c *Consumer) processEntry(ctx context.Context, d rueidis.DedicatedClient, entry StreamMessage) {
	handler, ok := c.handlers[entry.Type]
	if !ok {
		_ = d.Do(ctx, d.B().Xack().Key(StreamBlogEvents).Group(c.group).Id(entry.ID).Build()).Error()
		return
	}
	doneKey := DoneKeyPrefix + entry.ID
	// 幂等：同一 Stream 消息 ID 仅处理一次（7 天 TTL），重复投递直接 ACK。
	okNX, err := c.rds.Do(ctx, c.rds.B().Set().Key(doneKey).Value("1").Nx().ExSeconds(IdempotencyTTL).Build()).AsBool()
	if err == nil && !okNX {
		_ = d.Do(ctx, d.B().Xack().Key(StreamBlogEvents).Group(c.group).Id(entry.ID).Build()).Error()
		return
	}
	if err := handler(ctx, entry.Payload); err != nil {
		c.log.Warn("event handler failed", zap.String("type", entry.Type), zap.String("id", entry.ID), zap.Error(err))
		return
	}
	_ = d.Do(ctx, d.B().Xack().Key(StreamBlogEvents).Group(c.group).Id(entry.ID).Build()).Error()
}

func parseXReadGroup(resp rueidis.RedisResult) ([]StreamMessage, error) {
	arr, err := resp.ToArray()
	if err != nil {
		return nil, err
	}
	var out []StreamMessage
	for _, streamEntry := range arr {
		pairs, err := streamEntry.ToArray()
		if err != nil || len(pairs) < 2 {
			continue
		}
		msgs, err := pairs[1].ToArray()
		if err != nil {
			continue
		}
		for _, msg := range msgs {
			fields, err := msg.ToArray()
			if err != nil || len(fields) < 2 {
				continue
			}
			id, _ := fields[0].ToString()
			kv, err := fields[1].ToArray()
			if err != nil {
				continue
			}
			sm := StreamMessage{ID: id}
			for i := 0; i+1 < len(kv); i += 2 {
				k, _ := kv[i].ToString()
				v, _ := kv[i+1].ToString()
				switch k {
				case "type":
					sm.Type = v
				case "payload":
					sm.Payload = json.RawMessage(v)
				}
			}
			if sm.Type != "" {
				out = append(out, sm)
			}
		}
	}
	return out, nil
}
