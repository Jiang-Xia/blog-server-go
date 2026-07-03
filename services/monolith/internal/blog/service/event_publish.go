// event_publish 领域事件发布辅助，供各 blog service 复用。
package service

import (
	"context"

	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
	blogevent "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/event"
)

// domainEventPublisher 领域事件发布端口（*event.Publisher 实现）。
type domainEventPublisher interface {
	Publish(ctx context.Context, eventType string, payload interface{})
}

// publishSensitiveWordHit 命中敏感词且需扣 HP 时发布事件（对齐 Nest：拒绝前发布）。
func publishSensitiveWordHit(ctx context.Context, pub domainEventPublisher, uid, hpPenalty int) {
	if pub == nil || uid <= 0 || hpPenalty <= 0 {
		return
	}
	pub.Publish(ctx, blogevent.EventSensitiveWordHit, blogevent.SensitiveWordHitPayload{
		UID: uid, HpPenalty: hpPenalty,
	})
}

// publishArticleLifecycleEvents 文章状态变更后发布生命周期事件（对齐 Nest publishRagArticleEvents）。
func publishArticleLifecycleEvents(ctx context.Context, pub domainEventPublisher, article *ent.Article, prevStatus string, prevIsDelete bool) {
	if pub == nil || article == nil {
		return
	}
	payload := blogevent.ArticleLifecyclePayload{UID: article.UID, ArticleID: article.ID}
	if article.IsDelete && !prevIsDelete {
		pub.Publish(ctx, blogevent.EventArticleDeleted, payload)
		return
	}
	if article.Status == "publish" && !article.IsDelete {
		if prevStatus != "publish" {
			pub.Publish(ctx, blogevent.EventArticlePublished, payload)
		} else {
			pub.Publish(ctx, blogevent.EventArticleUpdated, payload)
		}
		return
	}
	if prevStatus == "publish" && article.Status != "publish" {
		pub.Publish(ctx, blogevent.EventArticleUnpublished, payload)
	}
}
