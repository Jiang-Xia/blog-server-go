// payloads 领域事件 payload 结构体，与 Nest blog.events.ts 对齐。
package event

// UserActionPayload 用户动作基础 payload。
type UserActionPayload struct {
	UID int `json:"uid"`
}

// SensitiveWordHitPayload 敏感词命中扣 HP。
type SensitiveWordHitPayload struct {
	UID       int `json:"uid"`
	HpPenalty int `json:"hpPenalty,omitempty"`
}

// ArticleLifecyclePayload 文章生命周期事件 payload（发布/更新/下架/删除）。
type ArticleLifecyclePayload struct {
	UID       int `json:"uid"`
	ArticleID int `json:"articleId"`
}

// ArticlePublishedPayload 文章发布事件 payload（scheduled_publish 等复用）。
type ArticlePublishedPayload = ArticleLifecyclePayload

// ArticleInteractionPayload 文章互动 payload。
type ArticleInteractionPayload struct {
	UID       int `json:"uid"`
	ArticleID int `json:"articleId"`
	AuthorUID int `json:"authorUid"`
}

// CommentCreatedPayload 评论创建事件。
type CommentCreatedPayload = ArticleInteractionPayload

// LikeCreatedPayload 点赞事件。
type LikeCreatedPayload struct {
	UID        int `json:"uid"`
	ArticleID  int `json:"articleId"`
	AuthorUID  int `json:"authorUid"`
	DailyLimit int `json:"dailyLimit,omitempty"`
}

// CollectCreatedPayload 收藏事件。
type CollectCreatedPayload struct {
	UID        int `json:"uid"`
	ArticleID  int `json:"articleId"`
	AuthorUID  int `json:"authorUid"`
	DailyLimit int `json:"dailyLimit,omitempty"`
}

// ReplyCreatedPayload 回复创建事件。
type ReplyCreatedPayload struct {
	UID int `json:"uid"`
}

// MsgboardCreatedPayload 留言创建事件。
type MsgboardCreatedPayload struct {
	UID int `json:"uid,omitempty"`
}

// ArticleViewedPayload 文章阅读事件。
type ArticleViewedPayload struct {
	ArticleID int `json:"articleId"`
	AuthorUID int `json:"authorUid"`
	ViewerUID int `json:"viewerUid,omitempty"`
}
