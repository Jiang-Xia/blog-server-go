// Package event RPG 域 Redis Stream 事件消费，驱动经验/任务/成就。
package event

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Jiang-Xia/blog-server-go/pkg/redisutil"
	blogevent "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/event"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/achievement"
	rpgconst "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/constants"
	rpgcore "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/core"
	rpglevel "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/level"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/quest"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/social"
)

// Handlers RPG 事件处理依赖。
type Handlers struct {
	Core        *rpgcore.RpgService
	Level       *rpglevel.LevelService
	Achievement *achievement.Service
	Quest       *quest.Service
	Reputation  *social.ReputationService
	Redis       *redisutil.Store
}

// RegisterRPGHandlers 向 event.Consumer 注册 RPG 侧 blog 事件 handler。
func RegisterRPGHandlers(c *blogevent.Consumer, h Handlers) {
	c.Register(blogevent.EventArticlePublished, h.onArticlePublished)
	c.Register(blogevent.EventCommentCreated, h.onCommentCreated)
	c.Register(blogevent.EventReplyCreated, h.onReplyCreated)
	c.Register(blogevent.EventMsgboardCreated, h.onMsgboardCreated)
	c.Register(blogevent.EventLikeCreated, h.onLikeCreated)
	c.Register(blogevent.EventCollectCreated, h.onCollectCreated)
	c.Register(blogevent.EventArticleViewed, h.onArticleViewed)
	c.Register(blogevent.EventArticleTipped, h.onArticleTipped)
	c.Register(blogevent.EventSeasonPosterShared, h.onSeasonPosterShared)
	c.Register(blogevent.EventUserRegistered, h.onUserRegistered)
	c.Register(blogevent.EventSensitiveWordHit, h.onSensitiveWordHit)
}

func (h Handlers) onArticlePublished(ctx context.Context, payload json.RawMessage) error {
	var p struct {
		UID       int `json:"uid"`
		ArticleID int `json:"articleId"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return err
	}
	_, _ = h.Level.AddExp(ctx, p.UID, 20, rpgcore.ExpReasonArticle, 0)
	if h.Achievement != nil {
		_ = h.Achievement.TrackProgress(ctx, p.UID, "article")
	}
	if h.Quest != nil {
		_ = h.Quest.TrackProgress(ctx, p.UID, "article")
	}
	if p.ArticleID > 0 && h.Reputation != nil {
		rep, _ := h.Reputation.GetReputation(ctx, p.UID)
		boost := h.Reputation.GetPublishExpBoostRate(rep)
		initialExp := int(float64(rpgconst.Economy.ArticlePublishBaseExp) * boost)
		_ = initialExp // 文章等级服务 Plan 09 后续接入
	}
	return nil
}

func (h Handlers) onCommentCreated(ctx context.Context, payload json.RawMessage) error {
	var p struct {
		UID int `json:"uid"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return err
	}
	_, _ = h.Level.AddExp(ctx, p.UID, 5, rpgcore.ExpReasonComment, 0)
	h.trackAchievement(ctx, p.UID, "comment")
	return h.trackQuest(ctx, p.UID, "comment")
}

func (h Handlers) onReplyCreated(ctx context.Context, payload json.RawMessage) error {
	var p struct {
		UID int `json:"uid"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return err
	}
	_, _ = h.Level.AddExp(ctx, p.UID, 5, rpgcore.ExpReasonReply, 0)
	h.trackAchievement(ctx, p.UID, "reply")
	return h.trackQuest(ctx, p.UID, "reply")
}

func (h Handlers) onMsgboardCreated(ctx context.Context, payload json.RawMessage) error {
	var p struct {
		UID int `json:"uid"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return err
	}
	_, _ = h.Level.AddExp(ctx, p.UID, 5, rpgcore.ExpReasonMsgboard, 0)
	h.trackAchievement(ctx, p.UID, "msgboard")
	return h.trackQuest(ctx, p.UID, "msgboard")
}

func (h Handlers) onLikeCreated(ctx context.Context, payload json.RawMessage) error {
	var p struct {
		UID        int `json:"uid"`
		DailyLimit int `json:"dailyLimit"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return err
	}
	limit := p.DailyLimit
	if limit <= 0 {
		limit = 10
	}
	_, _ = h.Level.AddExp(ctx, p.UID, 2, rpgcore.ExpReasonLike, limit)
	h.trackAchievement(ctx, p.UID, "like")
	return h.trackQuest(ctx, p.UID, "like")
}

func (h Handlers) onCollectCreated(ctx context.Context, payload json.RawMessage) error {
	var p struct {
		UID        int `json:"uid"`
		DailyLimit int `json:"dailyLimit"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return err
	}
	limit := p.DailyLimit
	if limit <= 0 {
		limit = 15
	}
	_, _ = h.Level.AddExp(ctx, p.UID, 3, rpgcore.ExpReasonCollect, limit)
	h.trackAchievement(ctx, p.UID, "collect")
	return h.trackQuest(ctx, p.UID, "collect")
}

func (h Handlers) onArticleViewed(ctx context.Context, payload json.RawMessage) error {
	var p struct {
		ArticleID int `json:"articleId"`
		ViewerUID int `json:"viewerUid"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return err
	}
	if h.Redis == nil {
		return nil
	}
	key := fmt.Sprintf("rpg:view:%d:%d:%s", p.ArticleID, p.ViewerUID, time.Now().Format("2006-01-02"))
	seen, _ := h.Redis.Get(ctx, key)
	if seen != "" {
		return nil
	}
	_ = h.Redis.Set(ctx, key, "1", 86400)
	return nil
}

func (h Handlers) onArticleTipped(ctx context.Context, payload json.RawMessage) error {
	var p struct {
		UID       int `json:"uid"`
		AuthorUID int `json:"authorUid"`
		Amount    int `json:"amount"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return err
	}
	if h.Reputation != nil {
		_, _ = h.Reputation.AddReputation(ctx, p.AuthorUID, (p.Amount+1)/2, "tip")
	}
	h.trackQuest(ctx, p.UID, "tip")
	h.trackAchievement(ctx, p.UID, "tip")
	h.trackAchievement(ctx, p.AuthorUID, "tip_received")
	return nil
}

func (h Handlers) onSeasonPosterShared(ctx context.Context, payload json.RawMessage) error {
	var p struct {
		UID int `json:"uid"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return err
	}
	h.trackAchievement(ctx, p.UID, "poster_share")
	return nil
}

func (h Handlers) onUserRegistered(ctx context.Context, payload json.RawMessage) error {
	var p struct {
		UID int `json:"uid"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return err
	}
	if h.Redis == nil {
		return nil
	}
	flagKey := fmt.Sprintf("rpg:register_bonus:%d", p.UID)
	ok, err := h.Redis.SetNX(ctx, flagKey, "1", 10*365*24*3600)
	if err != nil || !ok {
		return err
	}
	_, err = h.Core.GetOrCreateRpg(ctx, p.UID)
	return err
}

func (h Handlers) onSensitiveWordHit(ctx context.Context, payload json.RawMessage) error {
	var p struct {
		UID       int `json:"uid"`
		HpPenalty int `json:"hpPenalty"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return err
	}
	penalty := p.HpPenalty
	if penalty <= 0 {
		penalty = 10
	}
	rpg, err := h.Core.GetOrCreateRpg(ctx, p.UID)
	if err != nil {
		return err
	}
	rpg.LifeValue -= penalty
	if rpg.LifeValue < 0 {
		rpg.LifeValue = 0
		rpg.ZeroLifeCount++
	}
	rpg.SensitiveHitsCount++
	_, err = h.Core.SaveRpg(ctx, rpg)
	return err
}

func (h Handlers) trackAchievement(ctx context.Context, uid int, event string) {
	if h.Achievement != nil {
		_ = h.Achievement.TrackProgress(ctx, uid, event)
	}
}

func (h Handlers) trackQuest(ctx context.Context, uid int, action string) error {
	if h.Quest == nil {
		return nil
	}
	return h.Quest.TrackProgress(ctx, uid, action)
}
