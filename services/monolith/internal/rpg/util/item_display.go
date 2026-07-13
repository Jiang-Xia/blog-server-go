// item_display 背包/物品 C 端展示格式化（对齐 Nest item-config.util.ts）。
package util

import (
	"math"

	rpgconst "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/constants"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
)

// RoundExpBuffRate 经验倍率展示舍入，避免 JSON 浮点长尾。
func RoundExpBuffRate(rate float64) float64 {
	if rate <= 0 {
		return 1
	}
	return math.Round(rate*10000) / 10000
}

// FormatItemConfigForClient 将 rpg_item_config 转为 C 端 config 嵌套结构。
func FormatItemConfigForClient(cfg *ent.RpgItemConfig) map[string]interface{} {
	if cfg == nil {
		return nil
	}
	typeLabel, typeIcon := rpgconst.GetItemTypeDisplay(cfg.ItemType)
	rd := rpgconst.GetRarityDisplay(cfg.Rarity)
	item := map[string]interface{}{
		"code":          cfg.Code,
		"name":          cfg.Name,
		"itemType":      cfg.ItemType,
		"rarity":        cfg.Rarity,
		"description":   cfg.Description,
		"icon":          cfg.Icon,
		"category":      cfg.Category,
		"sort":          cfg.Sort,
		"itemTypeLabel": typeLabel,
		"itemTypeIcon":  typeIcon,
		"rarityLabel":   rd.Label,
		"rarityColor":   rd.Color,
		"rarityIcon":    rd.Icon,
	}
	if cfg.EffectJson != nil {
		item["effectJson"] = ParseEffectJSON(cfg.EffectJson)
	}
	return item
}

// FormatInventoryItemForClient 背包单条：sourceLabel + 嵌套 config。
func FormatInventoryItemForClient(row *ent.RpgUserInventory, cfg *ent.RpgItemConfig) map[string]interface{} {
	source := row.Source
	if source == "" {
		source = "system"
	}
	return map[string]interface{}{
		"id":          row.ID,
		"itemCode":    row.ItemCode,
		"quantity":    row.Quantity,
		"source":      source,
		"sourceLabel": rpgconst.GetItemSourceLabel(source),
		"acquiredAt":  row.AcquiredAt,
		"config":      FormatItemConfigForClient(cfg),
	}
}

// FormatActivitySummary C 端活动卡片（舍入 expBuffRate）。
func FormatActivitySummary(a *ent.RpgActivity) map[string]interface{} {
	if a == nil {
		return nil
	}
	return map[string]interface{}{
		"id":           a.ID,
		"code":         a.Code,
		"name":         a.Name,
		"description":  a.Description,
		"activityType": a.ActivityType,
		"startTime":    a.StartTime,
		"endTime":      a.EndTime,
		"expBuffRate":  RoundExpBuffRate(a.ExpBuffRate),
		"posterUrl":    a.PosterUrl,
		"active":       a.Active == 1,
	}
}
