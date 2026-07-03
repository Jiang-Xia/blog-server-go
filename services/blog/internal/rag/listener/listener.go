// listener RAG Stream 事件消费注册（增量索引）。
package listener

import (
	"context"
	"encoding/json"

	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/event"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/rag"
	"go.uber.org/zap"
)

const consumerGroupRAG = "rag-indexer"

// ArticleEventPayload 文章事件 payload。
type ArticleEventPayload struct {
	UID       int `json:"uid"`
	ArticleID int `json:"articleId"`
}

// UserLockedPayload 用户锁定事件。
type UserLockedPayload struct {
	UID int `json:"uid"`
}

// RegisterRAGHandlers 注册 RAG 增量索引 Stream 处理器。
func RegisterRAGHandlers(c *event.Consumer, mod *rag.Module, log *zap.Logger) {
	if mod == nil || mod.Indexer == nil {
		return
	}
	idx := mod.Indexer

	c.Register(event.EventArticlePublished, func(ctx context.Context, raw json.RawMessage) error {
		var p ArticleEventPayload
		if err := json.Unmarshal(raw, &p); err != nil {
			return nil
		}
		idx.IndexArticleByID(ctx, p.ArticleID)
		return nil
	})
	c.Register(event.EventArticleUpdated, func(ctx context.Context, raw json.RawMessage) error {
		var p ArticleEventPayload
		if err := json.Unmarshal(raw, &p); err != nil {
			return nil
		}
		idx.IndexArticleByID(ctx, p.ArticleID)
		return nil
	})
	c.Register(event.EventArticleUnpublished, func(ctx context.Context, raw json.RawMessage) error {
		var p ArticleEventPayload
		if err := json.Unmarshal(raw, &p); err != nil {
			return nil
		}
		return idx.RemoveArticleChunks(ctx, p.ArticleID)
	})
	c.Register(event.EventArticleDeleted, func(ctx context.Context, raw json.RawMessage) error {
		var p ArticleEventPayload
		if err := json.Unmarshal(raw, &p); err != nil {
			return nil
		}
		return idx.RemoveArticleChunks(ctx, p.ArticleID)
	})
	c.Register(event.EventUserLocked, func(ctx context.Context, raw json.RawMessage) error {
		var p UserLockedPayload
		if err := json.Unmarshal(raw, &p); err != nil {
			return nil
		}
		if p.UID > 0 {
			_, err := idx.PurgeAuthorArticles(ctx, p.UID)
			return err
		}
		return nil
	})
}

// ConsumerGroupRAG RAG 专用消费组名。
func ConsumerGroupRAG() string { return consumerGroupRAG }
