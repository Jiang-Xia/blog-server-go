// publisher 向 blog:events Stream 发布领域事件。
package event

import (
	"context"
	"encoding/json"

	"github.com/redis/rueidis"
	"go.uber.org/zap"
)

// Publisher 向 blog:events Stream 发布领域事件。
type Publisher struct {
	rds rueidis.Client
	log *zap.Logger
}

// NewPublisher 构造 Publisher。
func NewPublisher(rds rueidis.Client, log *zap.Logger) *Publisher {
	return &Publisher{rds: rds, log: log}
}

// Publish 发布事件；失败只记录日志，不阻断主流程。
func (p *Publisher) Publish(ctx context.Context, eventType string, payload interface{}) {
	raw, err := json.Marshal(payload)
	if err != nil {
		p.log.Warn("event publish marshal failed", zap.String("type", eventType), zap.Error(err))
		return
	}
	err = p.rds.Do(ctx,
		p.rds.B().Xadd().Key(StreamBlogEvents).Id("*").
			FieldValue().FieldValue("type", eventType).FieldValue("payload", string(raw)).
			Build(),
	).Error()
	if err != nil {
		p.log.Warn("event publish failed", zap.String("type", eventType), zap.Error(err))
	}
}

// StreamMessage Stream 单条消息。
type StreamMessage struct {
	ID      string
	Type    string
	Payload json.RawMessage
}

// PublishTest 发布测试事件（dev/冒烟）。
func (p *Publisher) PublishTest(ctx context.Context, eventType string, payload interface{}) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return p.rds.Do(ctx,
		p.rds.B().Xadd().Key(StreamBlogEvents).Id("*").
			FieldValue().FieldValue("type", eventType).FieldValue("payload", string(raw)).
			Build(),
	).Error()
}
