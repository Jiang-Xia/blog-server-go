// Package notify RPG WebSocket 推送编排：payload enrich、expGain 防抖。
package notify

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/Jiang-Xia/blog-server-go/pkg/redisutil"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/blog/ws"
	rpgconst "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/constants"
	rpgcore "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/core"
	rpgrepo "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/repo"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/util"
	userrepo "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/repo"
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

// ItemConfigReader 物品配置查询（itemGranted / petHatched enrich）。
type ItemConfigReader interface {
	FindItemConfigByCode(ctx context.Context, code string) (*ent.RpgItemConfig, error)
}

// OnlineUIDsProvider 返回当前在线用户 uid 列表（广播用）。
type OnlineUIDsProvider func() []uint64

// RpgNotifyService 封装 RPG WS 推送与 expGain 防抖。
type RpgNotifyService struct {
	pusher    ws.Pusher
	redis     *redisutil.Store
	users     *userrepo.UserRepo
	items     ItemConfigReader
	online    OnlineUIDsProvider
	expTimers sync.Map // uid -> *time.Timer
}

// NewRpgNotifyService 构造 RpgNotifyService；online 可为 nil（跳过广播）。
func NewRpgNotifyService(
	pusher ws.Pusher,
	redis *redisutil.Store,
	users *userrepo.UserRepo,
	online OnlineUIDsProvider,
) *RpgNotifyService {
	return &RpgNotifyService{pusher: pusher, redis: redis, users: users, online: online}
}

// SetItemConfigReader 延迟注入物品配置查询（避免 notify ↔ repo 构造顺序问题）。
func (s *RpgNotifyService) SetItemConfigReader(r ItemConfigReader) {
	s.items = r
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

// NotifyQuestComplete 推送任务完成（待领取）。
func (s *RpgNotifyService) NotifyQuestComplete(ctx context.Context, uid int, data QuestCompletePayload) {
	if s.pusher == nil {
		return
	}
	_ = s.pusher.PushToUser(ctx, uint64(uid), msgQuestComplete, 0, data)
}

// NotifyQuestReward 推送任务奖励已领取。
func (s *RpgNotifyService) NotifyQuestReward(ctx context.Context, uid int, data QuestRewardPayload) {
	if s.pusher == nil {
		return
	}
	_ = s.pusher.PushToUser(ctx, uint64(uid), msgQuestReward, 0, data)
}

// NotifyAchievementComplete 推送成就达成。
func (s *RpgNotifyService) NotifyAchievementComplete(ctx context.Context, uid int, data AchievementCompletePayload) {
	if s.pusher == nil {
		return
	}
	_ = s.pusher.PushToUser(ctx, uint64(uid), msgAchievementComplete, 0, data)
}

// NotifySocialReceived 推送社交互动；hpDelta≠0 时额外推 lifeChange。
func (s *RpgNotifyService) NotifySocialReceived(ctx context.Context, toUID int, data SocialReceivedPayload) {
	if s.pusher == nil {
		return
	}
	if data.FromNickname == "" && data.FromUID > 0 {
		data.FromNickname = s.resolveNickname(ctx, data.FromUID)
	}
	_ = s.pusher.PushToUser(ctx, uint64(toUID), msgSocialReceived, 0, data)
	if data.HpDelta != 0 {
		payload := LifeChangePayload{CurrentLife: data.CurrentLife}
		if data.HpDelta < 0 {
			d := -data.HpDelta
			payload.LifeDeducted = d
		} else {
			payload.LifeRecovered = &data.HpDelta
		}
		_ = s.pusher.PushToUser(ctx, uint64(toUID), msgLifeChange, 0, payload)
	}
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

// NotifyArticleLevelUp 推送文章升级（作者）。
func (s *RpgNotifyService) NotifyArticleLevelUp(ctx context.Context, authorUID int, data ArticleLevelUpPayload) {
	if s.pusher == nil {
		return
	}
	_ = s.pusher.PushToUser(ctx, uint64(authorUID), msgArticleLevelUp, 0, data)
}

// NotifyMasterpiece 推送文章晋升神作（作者）。
func (s *RpgNotifyService) NotifyMasterpiece(ctx context.Context, authorUID int, data MasterpiecePayload) {
	if s.pusher == nil {
		return
	}
	_ = s.pusher.PushToUser(ctx, uint64(authorUID), msgMasterpiece, 0, data)
}

// NotifyCurrencyChange 推送钻石余额变动。
func (s *RpgNotifyService) NotifyCurrencyChange(ctx context.Context, uid, delta, balance int, reason string) {
	if s.pusher == nil {
		return
	}
	_ = s.pusher.PushToUser(ctx, uint64(uid), msgCurrencyChange, 0, CurrencyChangePayload{
		Delta:       delta,
		Balance:     balance,
		Reason:      reason,
		ReasonLabel: rpgconst.GetItemSourceLabel(reason),
	})
}

// NotifyRechargeComplete 推送充值到账（带订单号）。
func (s *RpgNotifyService) NotifyRechargeComplete(ctx context.Context, uid int, payload RechargeCompletePayload) error {
	if s.pusher == nil {
		return nil
	}
	return s.pusher.PushToUser(ctx, uint64(uid), msgRechargeComplete, 0, payload)
}

// NotifyItemGranted 推送获得背包物品；跳过 level_up 与 buff 卷轴类。
func (s *RpgNotifyService) NotifyItemGranted(ctx context.Context, uid int, itemCode, source string, quantity int) {
	if s.pusher == nil || source == "level_up" || quantity <= 0 {
		return
	}
	var cfg *ent.RpgItemConfig
	if s.items != nil {
		cfg, _ = s.items.FindItemConfigByCode(ctx, itemCode)
	}
	if cfg != nil {
		effect := util.ParseEffectJSON(cfg.EffectJson)
		if grantType, _ := effect["grantType"].(string); grantType == "buff" {
			return
		}
	}
	payload := buildItemGrantedPayload(cfg, itemCode, source, quantity)
	_ = s.pusher.PushToUser(ctx, uint64(uid), msgItemGranted, 0, payload)
}

// NotifyLotteryTicketChange 推送抽奖券数量变动。
func (s *RpgNotifyService) NotifyLotteryTicketChange(ctx context.Context, uid, delta, total int, reason string) {
	if s.pusher == nil || delta == 0 {
		return
	}
	if reason == "" {
		reason = "system"
	}
	_ = s.pusher.PushToUser(ctx, uint64(uid), msgLotteryTicketChange, 0, LotteryTicketChangePayload{
		Delta:       delta,
		Total:       total,
		Reason:      reason,
		ReasonLabel: rpgconst.GetItemSourceLabel(reason),
	})
}

// NotifyPetHatched 推送宠物孵化/兑换成功。
func (s *RpgNotifyService) NotifyPetHatched(ctx context.Context, uid, petID int, petCode, name string) {
	if s.pusher == nil {
		return
	}
	rarity := rpgconst.GetRarityDisplay("common")
	if s.items != nil {
		if cfg, err := s.items.FindItemConfigByCode(ctx, petCode); err == nil && cfg != nil {
			rarity = rpgconst.GetRarityDisplay(cfg.Rarity)
			if name == "" {
				name = cfg.Name
			}
		}
	}
	_ = s.pusher.PushToUser(ctx, uint64(uid), msgPetHatched, 0, PetHatchedPayload{
		PetID:       petID,
		PetCode:     petCode,
		Name:        name,
		RarityLabel: rarity.Label,
		RarityColor: rarity.Color,
	})
}

// NotifyShieldUsed 推送护盾抵消敏感词扣血。
func (s *RpgNotifyService) NotifyShieldUsed(ctx context.Context, uid int) {
	if s.pusher == nil {
		return
	}
	_ = s.pusher.PushToUser(ctx, uint64(uid), msgShieldUsed, 0, ShieldUsedPayload{BuffName: "护盾"})
}

// NotifyBuffGranted 推送获得 Buff。
func (s *RpgNotifyService) NotifyBuffGranted(ctx context.Context, uid int, data BuffGrantedPayload) {
	if s.pusher == nil {
		return
	}
	_ = s.pusher.PushToUser(ctx, uint64(uid), msgBuffGranted, 0, data)
}

// NotifyBuffExpired 推送 Buff 自然过期。
func (s *RpgNotifyService) NotifyBuffExpired(ctx context.Context, uid int, data BuffExpiredPayload) {
	if s.pusher == nil {
		return
	}
	_ = s.pusher.PushToUser(ctx, uint64(uid), msgBuffExpired, 0, data)
}

// NotifyLifeChange 推送生命值变化（扣血场景）。
func (s *RpgNotifyService) NotifyLifeChange(ctx context.Context, uid, lifeDeducted, currentLife int) {
	if s.pusher == nil || lifeDeducted <= 0 {
		return
	}
	_ = s.pusher.PushToUser(ctx, uint64(uid), msgLifeChange, 0, LifeChangePayload{
		LifeDeducted: lifeDeducted,
		CurrentLife:  currentLife,
	})
}

// NotifyBanStatus 推送禁言/解封状态。
func (s *RpgNotifyService) NotifyBanStatus(ctx context.Context, uid int, banned bool, banEndTime *time.Time, banReason *string) {
	if s.pusher == nil {
		return
	}
	_ = s.pusher.PushToUser(ctx, uint64(uid), msgBanStatus, 0, BanStatusPayload{
		Banned:     banned,
		BanEndTime: banEndTime,
		BanReason:  banReason,
	})
}

// NotifyRankChange 推送排行榜名次变动（仅 Top10 且 1h 同维度去重）。
func (s *RpgNotifyService) NotifyRankChange(ctx context.Context, uid int, scoreType, period string, rank int, score float64) {
	if s.pusher == nil || rank <= 0 || rank > rankNotifyTopN {
		return
	}
	if s.redis != nil {
		dedupeKey := fmt.Sprintf("rpg:ws:rank:%d:%s:%s", uid, scoreType, period)
		ok, err := s.redis.SetNX(ctx, dedupeKey, "1", rankDedupeTTLSec)
		if err != nil || !ok {
			return
		}
	}
	_ = s.pusher.PushToUser(ctx, uint64(uid), msgRankChange, 0, RankChangePayload{
		Type:   scoreType,
		Period: period,
		Rank:   rank,
		Score:  score,
	})
}

// NotifyGuildEvent 向多个用户广播公会事件。
func (s *RpgNotifyService) NotifyGuildEvent(ctx context.Context, uids []int, data GuildEventPayload) {
	if s.pusher == nil || len(uids) == 0 {
		return
	}
	for _, uid := range uids {
		_ = s.pusher.PushToUser(ctx, uint64(uid), msgGuildEvent, 0, data)
	}
}

// NotifyWeatherBuff 推送连接上下文天气加成。
func (s *RpgNotifyService) NotifyWeatherBuff(ctx context.Context, uid int, data WeatherBuffPayload) {
	if s.pusher == nil {
		return
	}
	_ = s.pusher.PushToUser(ctx, uint64(uid), msgWeatherBuff, 0, data)
}

// BroadcastActivityUpdate 向在线用户广播活动变更（connect 类型跳过）。
func (s *RpgNotifyService) BroadcastActivityUpdate(ctx context.Context, payload ActivityUpdatePayload) error {
	if s.pusher == nil || payload.Type == "connect" || len(payload.Activities) == 0 {
		return nil
	}
	for _, uid := range s.onlineUIDs() {
		_ = s.pusher.PushToUser(ctx, uid, msgActivityUpdate, 0, payload)
	}
	return nil
}

func buildItemGrantedPayload(cfg *ent.RpgItemConfig, itemCode, source string, quantity int) ItemGrantedPayload {
	name := itemCode
	var rarityLabel, rarityColor, itemTypeLabel string
	if cfg != nil {
		name = cfg.Name
		rd := rpgconst.GetRarityDisplay(cfg.Rarity)
		rarityLabel = rd.Label
		rarityColor = rd.Color
		itemTypeLabel = rpgconst.GetItemTypeLabel(cfg.ItemType)
	}
	return ItemGrantedPayload{
		ItemCode:    itemCode,
		Quantity:    quantity,
		Source:      source,
		SourceLabel: rpgconst.GetItemSourceLabel(source),
		Config: ItemGrantedConfig{
			Name:          name,
			RarityLabel:   rarityLabel,
			RarityColor:   rarityColor,
			ItemTypeLabel: itemTypeLabel,
		},
	}
}

// BuildAchievementCompletePayload 从成就配置组装 WS payload。
func BuildAchievementCompletePayload(cfg *ent.RpgItemConfig, effect map[string]interface{}) AchievementCompletePayload {
	expReward := intFromEffect(effect["expReward"])
	rd := rpgconst.GetRarityDisplay(cfg.Rarity)
	return AchievementCompletePayload{
		Code:           cfg.Code,
		Name:           cfg.Name,
		ExpReward:      expReward,
		Rarity:         cfg.Rarity,
		RarityLabel:    rd.Label,
		RarityColor:    rd.Color,
		RarityIcon:     rd.Icon,
		CurrencyReward: intFromEffect(effect["currencyReward"]),
		TicketReward:   intFromEffect(effect["ticketReward"]),
		HpReward:       intFromEffect(effect["hpReward"]),
	}
}

func intFromEffect(v interface{}) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	default:
		return 0
	}
}

func (s *RpgNotifyService) onlineUIDs() []uint64 {
	if s.online != nil {
		return s.online()
	}
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

// 编译期断言 repo 实现 ItemConfigReader。
var _ ItemConfigReader = (*rpgrepo.RpgRepo)(nil)
