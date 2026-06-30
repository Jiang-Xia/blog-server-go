package event

import (
	"context"
	"encoding/json"

	"go.uber.org/zap"
)

// ArticlePublishedPayload 文章发布事件 payload。
type ArticlePublishedPayload struct {
	UID       int `json:"uid"`
	ArticleID int `json:"articleId"`
}

// RegisterBlogHandlers 注册 blog 域 Stream 消费者（骨架；RPG 事件 Plan 09 接入）。
func RegisterBlogHandlers(c *Consumer, log *zap.Logger) {
	c.Register(EventArticlePublished, func(ctx context.Context, raw json.RawMessage) error {
		var p ArticlePublishedPayload
		if err := json.Unmarshal(raw, &p); err != nil {
			log.Warn("article.published payload invalid", zap.Error(err))
			return nil // 解析失败 ACK 丢弃
		}
		log.Debug("article.published consumed",
			zap.Int("uid", p.UID),
			zap.Int("articleId", p.ArticleID),
		)
		// Plan 09 前：仅日志；后续可刷新统计缓存
		return nil
	})
}
