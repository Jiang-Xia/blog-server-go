// Package admin RPG 管理端写操作（成就/任务/奖池/物品/活动/公会/解封）。
package admin

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent/rpgitemconfig"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/itemasset"
)

// CreateAchievement 首次配置成就类型系统物品。
func (s *Service) CreateAchievement(ctx context.Context, body map[string]interface{}) (interface{}, error) {
	itemCode := strField(body, "itemCode")
	if itemCode == "" {
		itemCode = strField(body, "code")
	}
	if itemCode == "" {
		return nil, errcode.WithMessage(errcode.InvalidParam, "必须选择系统物品（itemCode）")
	}
	item, err := s.repo.FindItemConfigByCode(ctx, itemCode)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errcode.WithMessage(errcode.InvalidParam, "系统物品不存在: %s", itemCode)
		}
		return nil, err
	}
	if item.ItemType != "achievement" {
		return nil, errcode.WithMessage(errcode.InvalidParam, "只能选择成就类型的系统物品")
	}
	merged, err := mergeEffectJSON(item.EffectJson, body)
	if err != nil {
		return nil, err
	}
	if v, ok := merged["achievementConfigured"].(bool); ok && v {
		return nil, errcode.WithMessage(errcode.Conflict, "该成就已配置: %s", itemCode)
	}
	category := strField(body, "category")
	if category == "" {
		category = item.Category
	}
	sort := intField(body, "sort", item.Sort)
	if err := s.assertAchievementConfigMap(category, sort, merged); err != nil {
		return nil, err
	}
	merged["achievementConfigured"] = true
	effectStr, err := effectToRepoString(merged)
	if err != nil {
		return nil, err
	}
	patch := map[string]interface{}{rpgitemconfig.FieldEffectJson: *effectStr}
	if c := strField(body, "category"); c != "" {
		patch[rpgitemconfig.FieldCategory] = c
	}
	if _, ok := body["sort"]; ok {
		patch[rpgitemconfig.FieldSort] = intField(body, "sort", item.Sort)
	}
	if _, ok := body["active"]; ok {
		patch[rpgitemconfig.FieldActive] = boolToIntField(body, "active", item.Active)
	}
	if _, ok := body["isHidden"]; ok {
		patch[rpgitemconfig.FieldIsHidden] = boolToIntField(body, "isHidden", item.IsHidden)
	}
	if err := s.repo.UpdateItemConfig(ctx, item.ID, patch); err != nil {
		return nil, err
	}
	return s.repo.FindItemConfigByCode(ctx, itemCode)
}

// UpdateAchievement 更新成就配置。
func (s *Service) UpdateAchievement(ctx context.Context, id int, body map[string]interface{}) (interface{}, error) {
	existing, err := s.repo.FindItemConfigByID(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errcode.WithMessage(errcode.NotFound, "成就配置不存在")
		}
		return nil, err
	}
	if existing.ItemType != "achievement" {
		return nil, errcode.WithMessage(errcode.NotFound, "成就配置不存在")
	}
	merged, err := mergeEffectJSON(existing.EffectJson, body)
	if err != nil {
		return nil, err
	}
	merged["achievementConfigured"] = true
	if err := s.assertAchievementConfigMap(strField(body, "category"), intField(body, "sort", existing.Sort), merged); err != nil {
		return nil, err
	}
	patch := map[string]interface{}{}
	if c := strField(body, "category"); c != "" {
		patch[rpgitemconfig.FieldCategory] = c
	}
	if _, ok := body["sort"]; ok {
		patch[rpgitemconfig.FieldSort] = intField(body, "sort", existing.Sort)
	}
	if _, ok := body["active"]; ok {
		patch[rpgitemconfig.FieldActive] = boolToIntField(body, "active", existing.Active)
	}
	if _, ok := body["isHidden"]; ok {
		patch[rpgitemconfig.FieldIsHidden] = boolToIntField(body, "isHidden", existing.IsHidden)
	}
	if body["effectJson"] != nil || len(merged) > 0 {
		effectStr, err := effectToRepoString(merged)
		if err != nil {
			return nil, err
		}
		if effectStr != nil {
			patch[rpgitemconfig.FieldEffectJson] = *effectStr
		}
	}
	if len(patch) == 0 {
		return existing, nil
	}
	if err := s.repo.UpdateItemConfig(ctx, existing.ID, patch); err != nil {
		return nil, err
	}
	return s.repo.FindItemConfigByID(ctx, existing.ID)
}

// DeleteAchievement 删除成就配置行。
func (s *Service) DeleteAchievement(ctx context.Context, id int) (interface{}, error) {
	if err := s.repo.DeleteItemConfig(ctx, id); err != nil {
		return nil, err
	}
	return map[string]interface{}{"affected": 1}, nil
}

func (s *Service) assertAchievementConfigMap(category string, sort int, effect map[string]interface{}) error {
	if strings.TrimSpace(category) == "" {
		return errcode.WithMessage(errcode.InvalidParam, "分类不能为空")
	}
	if effect != nil {
		trackEvent, _ := effect["trackEvent"].(string)
		if strings.TrimSpace(trackEvent) == "" {
			return errcode.WithMessage(errcode.InvalidParam, "追踪事件不能为空")
		}
		switch maxProgress := effect["maxProgress"].(type) {
		case float64:
			if maxProgress < 1 {
				return errcode.WithMessage(errcode.InvalidParam, "达成次数须为不小于 1 的整数")
			}
		case int:
			if maxProgress < 1 {
				return errcode.WithMessage(errcode.InvalidParam, "达成次数须为不小于 1 的整数")
			}
		default:
			return errcode.WithMessage(errcode.InvalidParam, "达成次数须为不小于 1 的整数")
		}
		switch expReward := effect["expReward"].(type) {
		case float64:
			if expReward < 0 {
				return errcode.WithMessage(errcode.InvalidParam, "经验奖励须为不小于 0 的整数")
			}
		case int:
			if expReward < 0 {
				return errcode.WithMessage(errcode.InvalidParam, "经验奖励须为不小于 0 的整数")
			}
		}
	}
	if sort < 0 {
		return errcode.WithMessage(errcode.InvalidParam, "排序须为不小于 0 的整数")
	}
	return nil
}

// CreateQuestFromBody 从 JSON 创建任务。
func (s *Service) CreateQuestFromBody(ctx context.Context, body map[string]interface{}) (interface{}, error) {
	code := strField(body, "code")
	if code == "" {
		return nil, errcode.WithMessage(errcode.InvalidParam, "code 不能为空")
	}
	effect, _ := effectJSONString(body)
	row := &ent.RpgQuest{
		Code:           code,
		Name:           strField(body, "name"),
		Description:    strField(body, "description"),
		Type:           strField(body, "type"),
		QuestSubtype:   strField(body, "questSubtype"),
		TargetAction:   strField(body, "targetAction"),
		TargetCount:    intField(body, "targetCount", 1),
		ExpReward:      intField(body, "expReward", 10),
		HpReward:       intField(body, "hpReward", 0),
		CurrencyReward: intField(body, "currencyReward", 0),
		Sort:           intField(body, "sort", 10),
		Active:         boolToIntField(body, "active", 1),
		EffectJson:     effect,
	}
	if row.Type == "" {
		row.Type = "daily"
	}
	if row.QuestSubtype == "" {
		row.QuestSubtype = "daily"
	}
	return s.CreateQuest(ctx, row)
}

// UpdateQuestFromBody 从 JSON 更新任务。
func (s *Service) UpdateQuestFromBody(ctx context.Context, id int, body map[string]interface{}) (interface{}, error) {
	patch := map[string]interface{}{}
	for _, key := range []string{"name", "description", "targetCount", "expReward", "hpReward", "currencyReward", "sort", "active"} {
		if _, ok := body[key]; ok {
			switch key {
			case "name", "description":
				patch[key] = strField(body, key)
			case "targetCount", "expReward", "hpReward", "currencyReward", "sort":
				patch[key] = intField(body, key, 0)
			case "active":
				patch[key] = boolToIntField(body, key, 1)
			}
		}
	}
	if body["effectJson"] != nil {
		effect, err := effectJSONString(body)
		if err != nil {
			return nil, err
		}
		if effect != nil {
			patch["effectJson"] = *effect
		}
	}
	if len(patch) == 0 {
		return body, nil
	}
	if err := s.UpdateQuest(ctx, id, patch); err != nil {
		return nil, err
	}
	quests, _, err := s.repo.ListQuestsAdmin(ctx, 0, 10000)
	if err != nil {
		return nil, err
	}
	for _, q := range quests {
		if q.ID == id {
			return q, nil
		}
	}
	return patch, nil
}

// CreateLotteryPoolFromBody 创建奖池条目。
func (s *Service) CreateLotteryPoolFromBody(ctx context.Context, body map[string]interface{}) (interface{}, error) {
	itemCode := strField(body, "itemCode")
	if itemCode == "" {
		itemCode = strField(body, "code")
	}
	if itemCode == "" {
		return nil, errcode.WithMessage(errcode.InvalidParam, "必须选择系统物品（itemCode）")
	}
	item, err := s.repo.FindItemConfigByCode(ctx, itemCode)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errcode.WithMessage(errcode.InvalidParam, "系统物品不存在: %s", itemCode)
		}
		return nil, err
	}
	if _, err := s.repo.FindLotteryPoolByItemCode(ctx, itemCode); err == nil {
		return nil, errcode.WithMessage(errcode.Conflict, "该物品已在奖池中: %s", itemCode)
	} else if !ent.IsNotFound(err) {
		return nil, err
	}
	rarity := strField(body, "rarity")
	if rarity == "" {
		rarity = item.Rarity
	}
	row := &ent.RpgLotteryPool{
		ItemCode:    itemCode,
		Probability: floatField(body, "probability", 0),
		Rarity:      rarity,
		Sort:        intField(body, "sort", 10),
		Active:      boolToIntField(body, "active", 1),
	}
	return s.repo.CreateLotteryPool(ctx, row)
}

// UpdateLotteryPoolFromBody 更新奖池条目。
func (s *Service) UpdateLotteryPoolFromBody(ctx context.Context, id int, body map[string]interface{}) (interface{}, error) {
	patch := map[string]interface{}{}
	if _, ok := body["probability"]; ok {
		patch["probability"] = floatField(body, "probability", 0)
	}
	if _, ok := body["rarity"]; ok {
		patch["rarity"] = strField(body, "rarity")
	}
	if _, ok := body["sort"]; ok {
		patch["sort"] = intField(body, "sort", 10)
	}
	if _, ok := body["active"]; ok {
		patch["active"] = boolToIntField(body, "active", 1)
	}
	if len(patch) == 0 {
		return body, nil
	}
	if err := s.repo.UpdateLotteryPool(ctx, id, patch); err != nil {
		return nil, err
	}
	pools, _, err := s.repo.ListLotteryPoolAdmin(ctx, 0, 10000)
	if err != nil {
		return nil, err
	}
	for _, p := range pools {
		if p.ID == id {
			return p, nil
		}
	}
	return patch, nil
}

// DeleteLotteryPool 删除奖池条目。
func (s *Service) DeleteLotteryPool(ctx context.Context, id int) (interface{}, error) {
	if err := s.repo.DeleteLotteryPool(ctx, id); err != nil {
		return nil, err
	}
	return map[string]interface{}{"affected": 1}, nil
}

// UnbanUser 管理员解禁 RPG 用户。
func (s *Service) UnbanUser(ctx context.Context, uid int) (interface{}, error) {
	if uid <= 0 {
		return nil, errcode.WithMessage(errcode.InvalidParam, "无效用户")
	}
	if s.punishment == nil {
		return nil, errcode.WithMessage(errcode.InternalError, "惩罚模块未加载")
	}
	return s.punishment.AdminUnban(ctx, uid)
}

// CreateItemFromBody 创建系统物品。
func (s *Service) CreateItemFromBody(ctx context.Context, body map[string]interface{}) (interface{}, error) {
	code := strField(body, "code")
	if code == "" {
		return nil, errcode.WithMessage(errcode.InvalidParam, "code 不能为空")
	}
	effect, _ := effectJSONString(body)
	row := &ent.RpgItemConfig{
		Code:        code,
		Name:        strField(body, "name"),
		ItemType:    strField(body, "itemType"),
		Description: strField(body, "description"),
		Category:    strField(body, "category"),
		Icon:        strField(body, "icon"),
		Rarity:      strField(body, "rarity"),
		Sort:        intField(body, "sort", 10),
		Active:      boolToIntField(body, "active", 1),
		IsHidden:    boolToIntField(body, "isHidden", 0),
		EffectJson:  effect,
	}
	if row.Icon == "" {
		row.Icon = "default"
	}
	if row.Rarity == "" {
		row.Rarity = "common"
	}
	return s.CreateItem(ctx, row)
}

// UpdateItemFromBody 更新系统物品。
func (s *Service) UpdateItemFromBody(ctx context.Context, id int, body map[string]interface{}) (interface{}, error) {
	patch := map[string]interface{}{}
	for _, key := range []string{"name", "description", "itemType", "category", "icon", "rarity", "sort", "active", "isHidden"} {
		if _, ok := body[key]; ok {
			switch key {
			case "name", "description", "itemType", "category", "icon", "rarity":
				patch[key] = strField(body, key)
			case "sort", "active", "isHidden":
				if key == "sort" {
					patch[key] = intField(body, key, 10)
				} else {
					patch[key] = boolToIntField(body, key, 0)
				}
			}
		}
	}
	if body["effectJson"] != nil {
		effect, err := effectJSONString(body)
		if err != nil {
			return nil, err
		}
		if effect != nil {
			patch["effectJson"] = *effect
		}
	}
	if len(patch) == 0 {
		return body, nil
	}
	if err := s.UpdateItem(ctx, id, patch); err != nil {
		return nil, err
	}
	items, _, err := s.repo.ListItemConfigsAdmin(ctx, 0, 10000)
	if err != nil {
		return nil, err
	}
	for _, it := range items {
		if it.ID == id {
			return it, nil
		}
	}
	return patch, nil
}

// UploadItemAsset 上传物品 icon/bg。
func (s *Service) UploadItemAsset(ctx context.Context, icon, assetType string, data []byte, filename, contentType string) (interface{}, error) {
	iconKey, err := itemasset.SanitizeIconKey(icon)
	if err != nil {
		return nil, err
	}
	kind, err := itemasset.ParseKind(assetType)
	if err != nil {
		return nil, err
	}
	return itemasset.Save(s.uploadRoot, s.staticPrefix, kind, iconKey, data, filename, contentType)
}

// DeleteItemAsset 删除物品 icon/bg 磁盘文件。
func (s *Service) DeleteItemAsset(ctx context.Context, icon, assetType string) (interface{}, error) {
	iconKey, err := itemasset.SanitizeIconKey(icon)
	if err != nil {
		return nil, err
	}
	kind, err := itemasset.ParseKind(assetType)
	if err != nil {
		return nil, err
	}
	return itemasset.Delete(s.uploadRoot, kind, iconKey)
}

// CreateActivityFromBody 创建活动。
func (s *Service) CreateActivityFromBody(ctx context.Context, body map[string]interface{}) (interface{}, error) {
	code := strField(body, "code")
	if code == "" {
		return nil, errcode.WithMessage(errcode.InvalidParam, "code 不能为空")
	}
	start, err := parseTimeField(body, "startTime")
	if err != nil {
		return nil, errcode.WithMessage(errcode.InvalidParam, "%s", err.Error())
	}
	end, err := parseTimeField(body, "endTime")
	if err != nil {
		return nil, errcode.WithMessage(errcode.InvalidParam, "%s", err.Error())
	}
	effect, _ := effectJSONString(body)
	row := &ent.RpgActivity{
		Code:         code,
		Name:         strField(body, "name"),
		Description:  strField(body, "description"),
		ActivityType: strField(body, "activityType"),
		PosterUrl:    strField(body, "posterUrl"),
		StartTime:    start,
		EndTime:      end,
		ExpBuffRate:  floatField(body, "expBuffRate", 1),
		Active:       boolToIntField(body, "active", 1),
		EffectJson:   effect,
	}
	if row.ActivityType == "" {
		row.ActivityType = "event"
	}
	return s.repo.CreateActivity(ctx, row)
}

// DeleteGuild 删除公会。
func (s *Service) DeleteGuild(ctx context.Context, id int) (interface{}, error) {
	if err := s.repo.DeleteGuild(ctx, id); err != nil {
		return nil, err
	}
	return map[string]interface{}{"affected": 1}, nil
}

// ListGuildMembers 公会成员详情。
func (s *Service) ListGuildMembers(ctx context.Context, guildID int) (interface{}, error) {
	return s.guild.GetDetail(ctx, guildID, 0)
}

// RemoveGuildMember 管理端移除公会成员。
func (s *Service) RemoveGuildMember(ctx context.Context, guildID, uid int) (interface{}, error) {
	return s.guild.AdminRemoveMember(ctx, guildID, uid)
}

// 避免未使用 import
var _ = json.Marshal

