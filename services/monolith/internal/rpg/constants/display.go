// Package constants RPG 展示文案（序列化进 WS/API，对齐 Nest item-display.ts）。
package constants

// RarityDisplay 稀有度展示。
type RarityDisplay struct {
	Color string
	Label string
	Icon  string
}

var rarityDisplay = map[string]RarityDisplay{
	"common":    {Color: "#c8d4e0", Label: "普通", Icon: "⚪"},
	"rare":      {Color: "#22c55e", Label: "稀有", Icon: "🟢"},
	"epic":      {Color: "#8b5cf6", Label: "史诗", Icon: "🟣"},
	"legendary": {Color: "#f59e0b", Label: "传说", Icon: "🟡"},
}

var itemTypeDisplay = map[string]struct {
	Label string
	Icon  string
}{
	"title":        {Label: "称号", Icon: "🏅"},
	"avatar_frame": {Label: "头像框", Icon: "🖼️"},
	"pet":          {Label: "宠物", Icon: "🐾"},
	"equipment":    {Label: "装备", Icon: "⚔️"},
	"achievement":  {Label: "成就", Icon: "🏆"},
	"buff":         {Label: "增益", Icon: "✨"},
	"currency":     {Label: "钻石", Icon: "💎"},
	"consumable":   {Label: "消耗品", Icon: "🧪"},
}

var itemSourceDisplay = map[string]string{
	"level_up":             "等级奖励",
	"lottery":              "抽奖",
	"lottery_reward":       "抽奖奖励",
	"quest":                "任务",
	"admin":                "管理员发放",
	"system":               "系统",
	"reward":               "奖励",
	"egg":                  "扔鸡蛋",
	"flower":               "送鲜花",
	"cheer":                "加油",
	"tip":                  "打赏",
	"register_bonus":       "注册奖励",
	"seed":                 "系统发放",
	"tip_received":         "收到打赏",
	"pet_exchange":         "宠物兑换",
	"sign_in":              "签到奖励",
	"consume":              "消耗",
	"all_quests_completed": "每日任务全完成",
	"recharge":             "充值",
	"admin_recharge":       "管理员充值",
	"achievement":          "成就",
}

// GetRarityDisplay 返回稀有度展示；未知 rarity 回退为原值。
func GetRarityDisplay(rarity string) RarityDisplay {
	if d, ok := rarityDisplay[rarity]; ok {
		return d
	}
	return RarityDisplay{Color: "#c8d4e0", Label: rarity, Icon: "⚪"}
}

// GetItemTypeLabel 返回物品类型中文标签。
func GetItemTypeLabel(itemType string) string {
	if d, ok := itemTypeDisplay[itemType]; ok {
		return d.Label
	}
	return itemType
}

// GetItemSourceLabel 返回背包来源 code 的中文展示。
func GetItemSourceLabel(source string) string {
	if l, ok := itemSourceDisplay[source]; ok {
		return l
	}
	return source
}
