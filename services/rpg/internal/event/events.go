// Package event Redis Stream 领域事件总线，与 Nest blog:events 对齐。
package event

const (
	// StreamBlogEvents Redis Stream 名。
	StreamBlogEvents = "blog:events"
	// ConsumerGroupBlog 消费组名（blog 域消费者）。
	ConsumerGroupBlog = "blog-handlers"
	// ConsumerGroupRPG 消费组名（RPG 域，Plan 09 接入）。
	ConsumerGroupRPG = "rpg-handlers"
	// DoneKeyPrefix 幂等标记前缀。
	DoneKeyPrefix = "blog:event:done:"
	// IdempotencyTTL 幂等标记 TTL（7 天）。
	IdempotencyTTL = 7 * 24 * 3600
	// StreamMaxLen Stream 近似裁剪长度。
	StreamMaxLen = 10000
)

// 事件名常量，与 Nest BlogEvents 对齐。
const (
	EventArticlePublished   = "blog.article.published"
	EventCommentCreated     = "blog.comment.created"
	EventReplyCreated       = "blog.reply.created"
	EventMsgboardCreated    = "blog.msgboard.created"
	EventLikeCreated        = "blog.like.created"
	EventCollectCreated     = "blog.collect.created"
	EventSensitiveWordHit   = "blog.sensitive-word.hit"
	EventArticleViewed      = "blog.article.viewed"
	EventArticleTipped      = "blog.article.tipped"
	EventSeasonPosterShared = "blog.season.poster.shared"
	EventSocialInteract     = "blog.social.interact"
	EventUserRegistered     = "blog.user.registered"
	EventArticleUpdated     = "blog.article.updated"
	EventArticleUnpublished = "blog.article.unpublished"
	EventArticleDeleted     = "blog.article.deleted"
	EventUserLocked         = "blog.user.locked"
)
