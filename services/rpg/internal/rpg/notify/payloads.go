// Package notify WS 事件名与 payload 契约（对齐 Nest ws-events.ts）。
package notify

import "time"

const (
	msgLevelUp              = "levelUp"
	msgExpGain              = "expGain"
	msgTipReceived          = "tipReceived"
	msgRechargeComplete     = "rechargeComplete"
	msgActivityUpdate       = "activityUpdate"
	msgShieldUsed           = "shieldUsed"
	msgBanStatus            = "banStatus"
	msgLifeChange           = "lifeChange"
	msgQuestComplete        = "questComplete"
	msgQuestReward          = "questReward"
	msgAchievementComplete  = "achievementComplete"
	msgSocialReceived       = "socialReceived"
	msgArticleLevelUp       = "articleLevelUp"
	msgMasterpiece          = "masterpiece"
	msgCurrencyChange       = "currencyChange"
	msgItemGranted          = "itemGranted"
	msgLotteryTicketChange  = "lotteryTicketChange"
	msgPetHatched           = "petHatched"
	msgBuffGranted          = "buffGranted"
	msgBuffExpired          = "buffExpired"
	msgRankChange           = "rankChange"
	msgGuildEvent           = "guildEvent"
	msgWeatherBuff          = "weatherBuff"

	expGainDebounceSec = 8
	rankNotifyTopN     = 10
	rankDedupeTTLSec   = 3600
)

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
	ArticleTitle string `json:"articleTitle,omitempty"`
	Amount       int    `json:"amount"`
	Balance      *int   `json:"balance,omitempty"`
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

// BanStatusPayload 禁言状态 WS payload。
type BanStatusPayload struct {
	Banned     bool       `json:"banned"`
	BanEndTime *time.Time `json:"banEndTime"`
	BanReason  *string    `json:"banReason"`
}

// LifeChangePayload 生命值变化 WS payload。
type LifeChangePayload struct {
	LifeDeducted  int  `json:"lifeDeducted"`
	CurrentLife   int  `json:"currentLife"`
	LifeRecovered *int `json:"lifeRecovered,omitempty"`
}

// ShieldUsedPayload 护盾抵消 WS payload。
type ShieldUsedPayload struct {
	BuffName string `json:"buffName"`
}

// QuestCompletePayload 任务完成待领取。
type QuestCompletePayload struct {
	QuestCode  string `json:"questCode"`
	QuestName  string `json:"questName"`
	ExpReward  int    `json:"expReward"`
	HpReward   *int   `json:"hpReward,omitempty"`
}

// QuestRewardPayload 任务奖励已领取。
type QuestRewardPayload struct {
	QuestCode string `json:"questCode"`
	QuestName string `json:"questName"`
	ExpReward int    `json:"expReward"`
}

// AchievementCompletePayload 成就达成。
type AchievementCompletePayload struct {
	Code            string `json:"code"`
	Name            string `json:"name"`
	ExpReward       int    `json:"expReward"`
	Rarity          string `json:"rarity,omitempty"`
	RarityLabel     string `json:"rarityLabel,omitempty"`
	RarityColor     string `json:"rarityColor,omitempty"`
	RarityIcon      string `json:"rarityIcon,omitempty"`
	CurrencyReward  int    `json:"currencyReward,omitempty"`
	TicketReward    int    `json:"ticketReward,omitempty"`
	HpReward        int    `json:"hpReward,omitempty"`
}

// SocialReceivedPayload 收到社交互动。
type SocialReceivedPayload struct {
	FromUID          int    `json:"fromUid"`
	FromNickname     string `json:"fromNickname"`
	Action           string `json:"action"`
	HpDelta          int    `json:"hpDelta"`
	CurrentLife      int    `json:"currentLife"`
	ReputationDelta  int    `json:"reputationDelta"`
}

// ArticleLevelUpPayload 文章等级提升。
type ArticleLevelUpPayload struct {
	ArticleID    int    `json:"articleId"`
	ArticleTitle string `json:"articleTitle"`
	OldLevel     int    `json:"oldLevel"`
	NewLevel     int    `json:"newLevel"`
}

// MasterpiecePayload 文章晋升神作。
type MasterpiecePayload struct {
	ArticleID    int    `json:"articleId"`
	ArticleTitle string `json:"articleTitle"`
}

// CurrencyChangePayload 钻石余额变动。
type CurrencyChangePayload struct {
	Delta       int    `json:"delta"`
	Balance     int    `json:"balance"`
	Reason      string `json:"reason"`
	ReasonLabel string `json:"reasonLabel"`
}

// ItemGrantedConfig 物品展示摘要。
type ItemGrantedConfig struct {
	Name          string `json:"name"`
	RarityLabel   string `json:"rarityLabel,omitempty"`
	RarityColor   string `json:"rarityColor,omitempty"`
	ItemTypeLabel string `json:"itemTypeLabel,omitempty"`
}

// ItemGrantedPayload 获得背包物品。
type ItemGrantedPayload struct {
	ItemCode    string            `json:"itemCode"`
	Quantity    int               `json:"quantity"`
	Source      string            `json:"source"`
	SourceLabel string            `json:"sourceLabel"`
	Config      ItemGrantedConfig `json:"config"`
}

// LotteryTicketChangePayload 抽奖券数量变动。
type LotteryTicketChangePayload struct {
	Delta       int    `json:"delta"`
	Total       int    `json:"total"`
	Reason      string `json:"reason"`
	ReasonLabel string `json:"reasonLabel"`
}

// PetHatchedPayload 宠物孵化成功。
type PetHatchedPayload struct {
	PetID       int    `json:"petId"`
	PetCode     string `json:"petCode"`
	Name        string `json:"name"`
	RarityLabel string `json:"rarityLabel"`
	RarityColor string `json:"rarityColor"`
}

// BuffGrantedPayload 获得 Buff。
type BuffGrantedPayload struct {
	Code        string     `json:"code"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	ExpireAt    *time.Time `json:"expireAt,omitempty"`
}

// BuffExpiredPayload Buff 自然过期。
type BuffExpiredPayload struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// RankChangePayload 排行榜名次变化。
type RankChangePayload struct {
	Type   string  `json:"type"`
	Period string  `json:"period"`
	Rank   int     `json:"rank"`
	Score  float64 `json:"score"`
}

// GuildEventPayload 公会事件。
type GuildEventPayload struct {
	Type      string `json:"type"`
	GuildID   int    `json:"guildId"`
	GuildName string `json:"guildName"`
	UID       int    `json:"uid"`
	Nickname  string `json:"nickname"`
}

// WeatherBuffPayload 连接时天气加成。
type WeatherBuffPayload struct {
	Label    string  `json:"label"`
	ExpBoost float64 `json:"expBoost"`
	Weather  string  `json:"weather"`
}
