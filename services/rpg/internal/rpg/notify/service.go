// Package notify RPG WebSocket 推送编排：payload enrich、expGain 防抖。
package notify

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/Jiang-Xia/blog-server-go/pkg/redisutil"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/internal/userport"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/internal/wspush"
	rpgcore "github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/core"
)

const (
	msgLevelUp          = "levelUp"
	msgExpGain          = "expGain"
	msgTipReceived      = "tipReceived"
	msgRechargeComplete = "rechargeComplete"
	msgActivityUpdate   = "activityUpdate"

	expGainDebounceSec = 8
)

var expReasonLabel = map[string]string{
	"sign_in":     "签到",
	"comment":     "评论",
	"msgboard":    "留言",
	"reply":       "回复",
	"article":     "发文",
	"like":        "点赞",
	"collect":     "收藏",
	"quest":       "任务",
	"achievement": "成就",
	"lottery":     "抽奖",
}

// ExpGainPayload expGain 合并推送 payload。
type ExpGainPayload struct {
	Amount       int      `json:"amount"`
	Reasons      []string `json:"reasons"`
	ReasonLabels []string `json:"reasonLabels"`
}

// TipReceivedPayload 收到打赏 payload。
type TipReceivedPayload struct {
	FromUID      int    `json:"fromUid"`
	FromNickname string `json:"fromNickname"`
	ArticleID    int    `json:"articleId"`
	Amount       int    `json:"amount"`
}

// RechargeCompletePayload 充值到账 payload。
type RechargeCompletePayload struct {
	OutTradeNo string  `json:"outTradeNo"`
	Diamonds   int     `json:"diamonds"`
	Balance    int     `json:"balance"`
	AmountYuan float64 `json:"amountYuan"`
}

// ActivityItem 活动摘要。
type ActivityItem struct {
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	ExpBuffRate float64 `json:"expBuffRate,omitempty"`
}

// ActivityUpdatePayload 活动变更广播 payload。
type ActivityUpdatePayload struct {
	Type       string         `json:"type"`
	Activities []ActivityItem `json:"activities"`
}

// RpgNotifyService 封装 RPG WS 推送与 expGain 防抖。
type RpgNotifyService struct {
	pusher   wspush.Pusher
	redis    *redisutil.Store
	users    userport.UserReader
	expTimers sync.Map // uid -> *time.Timer
}

// NewRpgNotifyService 构造 RpgNotifyService。
func NewRpgNotifyService(
	pusher wspush.Pusher,
	redis *redisutil.Store,
	users userport.UserReader,
) *RpgNotifyService {
	return &RpgNotifyService{pusher: pusher, redis: redis, users: users}
}

// NotifyLevelUp 推送用户升级事件。
func (s *RpgNotifyService) NotifyLevelUp(ctx context.Context, uid int, result *rpgcore.LevelUpResult) {
	if s.pusher == nil || result == nil {
		return
	}
	_ = s.pusher.PushToUser(ctx, uint64(uid), msgLevelUp, 0, result)
}

// NotifyExpGain 合并推送经验获得（8s 窗口内累加 amount/reasons）。
func (s *RpgNotifyService) NotifyExpGain(ctx context.Context, uid, amount int, reason string) {
	if s.pusher == nil || amount <= 0 {
		return
	}
	go func() {
		if err := s.accumulateExpGain(context.Background(), uid, amount, reason); err != nil {
			_ = err
		}
	}()
	_ = ctx
}

// NotifyTipReceived 推送收到文章打赏（作者侧）。
func (s *RpgNotifyService) NotifyTipReceived(ctx context.Context, authorUID int, data TipReceivedPayload) error {
	if s.pusher == nil {
		return nil
	}
	if data.FromNickname == "" && data.FromUID > 0 && s.users != nil {
		data.FromNickname = s.resolveNickname(ctx, data.FromUID)
	}
	return s.pusher.PushToUser(ctx, uint64(authorUID), msgTipReceived, 0, data)
}

// NotifyRechargeComplete 推送充值到账（带订单号）。
func (s *RpgNotifyService) NotifyRechargeComplete(ctx context.Context, uid int, payload RechargeCompletePayload) error {
	if s.pusher == nil {
		return nil
	}
	return s.pusher.PushToUser(ctx, uint64(uid), msgRechargeComplete, 0, payload)
}

// BroadcastActivityUpdate 向在线用户广播活动变更（rpg 独立进程无 Hub，跳过广播）。
func (s *RpgNotifyService) BroadcastActivityUpdate(ctx context.Context, payload ActivityUpdatePayload) error {
	_ = ctx
	_ = payload
	return nil
}

func (s *RpgNotifyService) onlineUIDs() []uint64 {
	return nil
}

func (s *RpgNotifyService) resolveNickname(ctx context.Context, uid int) string {
	if s.users == nil {
		return fmt.Sprintf("用户%d", uid)
	}
	user, err := s.users.FindByID(ctx, uid)
	if err != nil || user == nil {
		return fmt.Sprintf("用户%d", uid)
	}
	if user.Nickname != "" {
		return user.Nickname
	}
	if user.Username != nil && *user.Username != "" {
		return *user.Username
	}
	return fmt.Sprintf("用户%d", uid)
}

func (s *RpgNotifyService) accumulateExpGain(ctx context.Context, uid, amount int, reason string) error {
	if s.redis == nil {
		return s.flushExpGain(ctx, uid, ExpGainPayload{
			Amount: amount, Reasons: []string{reason}, ReasonLabels: []string{labelReason(reason)},
		})
	}

	dataKey := fmt.Sprintf("rpg:ws:exp:%d", uid)
	raw, err := s.redis.Get(ctx, dataKey)
	if err != nil {
		return err
	}
	var pending ExpGainPayload
	if raw != "" {
		_ = json.Unmarshal([]byte(raw), &pending)
	}
	pending.Amount += amount
	if !containsStr(pending.Reasons, reason) {
		pending.Reasons = append(pending.Reasons, reason)
		pending.ReasonLabels = append(pending.ReasonLabels, labelReason(reason))
	}
	buf, _ := json.Marshal(pending)
	if err := s.redis.Set(ctx, dataKey, string(buf), expGainDebounceSec+2); err != nil {
		return err
	}

	lockKey := fmt.Sprintf("rpg:ws:exp:lock:%d", uid)
	acquired, err := s.redis.SetNX(ctx, lockKey, "1", expGainDebounceSec)
	if err != nil {
		return err
	}
	if !acquired {
		return nil
	}

	if old, ok := s.expTimers.Load(uid); ok {
		if t, ok := old.(*time.Timer); ok {
			t.Stop()
		}
	}
	timer := time.AfterFunc(time.Duration(expGainDebounceSec)*time.Second, func() {
		s.expTimers.Delete(uid)
		_ = s.flushExpGain(context.Background(), uid, ExpGainPayload{})
	})
	s.expTimers.Store(uid, timer)
	return nil
}

func (s *RpgNotifyService) flushExpGain(ctx context.Context, uid int, direct ExpGainPayload) error {
	if s.pusher == nil {
		return nil
	}
	payload := direct
	if s.redis != nil && direct.Amount == 0 {
		dataKey := fmt.Sprintf("rpg:ws:exp:%d", uid)
		raw, err := s.redis.Get(ctx, dataKey)
		if err != nil || raw == "" {
			return err
		}
		_ = s.redis.Del(ctx, dataKey)
		if err := json.Unmarshal([]byte(raw), &payload); err != nil {
			return err
		}
	}
	if payload.Amount <= 0 {
		return nil
	}
	return s.pusher.PushToUser(ctx, uint64(uid), msgExpGain, 0, payload)
}

func labelReason(reason string) string {
	if l, ok := expReasonLabel[reason]; ok {
		return l
	}
	return reason
}

func containsStr(ss []string, v string) bool {
	for _, s := range ss {
		if s == v {
			return true
		}
	}
	return false
}
