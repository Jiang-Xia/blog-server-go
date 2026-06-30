package core

// ExpReason 经验增加原因（写入 Redis 每日上限分组与 WS 展示）。
type ExpReason string

const (
	ExpReasonSignIn      ExpReason = "sign_in"
	ExpReasonComment     ExpReason = "comment"
	ExpReasonMsgboard    ExpReason = "msgboard"
	ExpReasonReply       ExpReason = "reply"
	ExpReasonArticle     ExpReason = "article"
	ExpReasonLike        ExpReason = "like"
	ExpReasonCollect     ExpReason = "collect"
	ExpReasonQuest       ExpReason = "quest"
	ExpReasonAchievement ExpReason = "achievement"
	ExpReasonLottery     ExpReason = "lottery"
)

// LevelReward 等级奖励 C 端展示结构。
type LevelReward struct {
	Level          int    `json:"level"`
	CurrencyReward int    `json:"currencyReward,omitempty"`
	AvatarFrame    string `json:"avatarFrame,omitempty"`
	Title          string `json:"title,omitempty"`
}

// LevelUpResult 升级事件结果。
type LevelUpResult struct {
	OldLevel        int           `json:"oldLevel"`
	NewLevel        int           `json:"newLevel"`
	UnlockedRewards []LevelReward `json:"unlockedRewards"`
}

// ExpProgress 当前等级经验进度。
type ExpProgress struct {
	Current  int `json:"current"`
	Required int `json:"required"`
	Percent  int `json:"percent"`
}
