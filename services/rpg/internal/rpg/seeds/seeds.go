// Package seeds 预定义 RPG 种子数据与幂等 upsert 辅助。
package seeds

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Jiang-Xia/blog-server-go/services/rpg/ent"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/ent/rpgitemconfig"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/ent/rpglotterypool"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/ent/rpgquest"
	rpgrepo "github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/repo"
	"go.uber.org/zap"
)

// ItemConfigSeed 物品配置种子。
type ItemConfigSeed struct {
	Code        string
	ItemType    string
	Name        string
	Description string
	Category    string
	Icon        string
	Rarity      string
	Sort        int
	Active      int
	IsHidden    int
	Effect      map[string]interface{}
}

// QuestSeed 任务种子。
type QuestSeed struct {
	Code           string
	Name           string
	Description    string
	Type           string
	QuestSubtype   string
	TargetAction   string
	TargetCount    int
	ExpReward      int
	HpReward       int
	CurrencyReward int
	Sort           int
	Active         int
	Effect         map[string]interface{}
}

// LotteryPoolSeed 抽奖池种子。
type LotteryPoolSeed struct {
	ItemCode    string
	Probability float64
	Rarity      string
	Sort        int
	Active      int
}

// ActivitySeed 活动种子。
type ActivitySeed struct {
	Code         string
	Name         string
	Description  string
	ActivityType string
	StartTime    string
	EndTime      string
	ExpBuffRate  float64
	PosterURL    string
	Active       int
}

// CurrencyItemDef 钻石物品定义。
var CurrencyItemDef = ItemConfigSeed{
	Code:        "currency",
	ItemType:    "currency",
	Name:        "钻石",
	Description: "通用货币，可用于兑换宠物、抽奖、打赏等",
	Category:    "currency",
	Icon:        "diamond",
	Rarity:      "common",
	Sort:        1,
	Active:      1,
}

// PredefinedCosmeticSeeds 装扮类种子（节选，完整列表可迭代扩展）。
var PredefinedCosmeticSeeds = []ItemConfigSeed{
	{Code: "frame_a", ItemType: "avatar_frame", Name: "青铜框", Rarity: "common", Sort: 10, Active: 1, Effect: map[string]interface{}{"color": "#cd7f32"}},
	{Code: "frame_b", ItemType: "avatar_frame", Name: "白银框", Rarity: "rare", Sort: 11, Active: 1, Effect: map[string]interface{}{"color": "#c0c0c0"}},
	{Code: "bronze_master", ItemType: "title", Name: "青铜大师", Rarity: "rare", Sort: 20, Active: 1},
	{Code: "silver_master", ItemType: "title", Name: "白银大师", Rarity: "epic", Sort: 21, Active: 1},
}

// PredefinedLotteryItems 抽奖奖品种子（对齐 Nest DEFAULT_LOTTERY_ITEMS）。
var PredefinedLotteryItems = []ItemConfigSeed{
	{Code: "exp_5", ItemType: "consumable", Name: "经验碎片", Icon: "exp", Effect: map[string]interface{}{"grantType": "exp", "amount": 5}},
	{Code: "exp_10", ItemType: "consumable", Name: "经验小块", Icon: "exp", Effect: map[string]interface{}{"grantType": "exp", "amount": 10}},
	{Code: "ticket_1", ItemType: "consumable", Name: "抽奖券碎片", Icon: "ticket", Effect: map[string]interface{}{"grantType": "ticket", "amount": 1}},
	{Code: "exp_25", ItemType: "consumable", Name: "经验中块", Icon: "exp", Effect: map[string]interface{}{"grantType": "exp", "amount": 25}},
	{Code: "buff_exp_small", ItemType: "consumable", Name: "经验微增卷轴", Icon: "scroll", Effect: map[string]interface{}{"grantType": "buff", "buffCode": "exp_boost_small"}},
	{Code: "ticket_3", ItemType: "consumable", Name: "抽奖券小包", Icon: "ticket", Effect: map[string]interface{}{"grantType": "ticket", "amount": 3}},
	{Code: "exp_50", ItemType: "consumable", Name: "经验大块", Icon: "exp", Effect: map[string]interface{}{"grantType": "exp", "amount": 50}},
	{Code: "buff_exp_medium", ItemType: "consumable", Name: "经验大增卷轴", Icon: "scroll", Effect: map[string]interface{}{"grantType": "buff", "buffCode": "exp_boost_medium"}},
	{Code: "buff_shield", ItemType: "consumable", Name: "护盾卷轴", Icon: "shield", Effect: map[string]interface{}{"grantType": "buff", "buffCode": "shield_basic"}},
	{Code: "exp_100", ItemType: "consumable", Name: "经验宝珠", Icon: "gem", Effect: map[string]interface{}{"grantType": "exp", "amount": 100}},
	{Code: "ticket_10", ItemType: "consumable", Name: "抽奖券宝箱", Icon: "ticket", Effect: map[string]interface{}{"grantType": "ticket", "amount": 10}},
	{Code: "lottery_title_writer", ItemType: "title", Name: "抽奖称号·作家", Category: "lottery", Rarity: "rare", Sort: 50, Active: 1},
	{Code: "lottery_frame_star", ItemType: "avatar_frame", Name: "抽奖头像框·星芒", Category: "lottery", Rarity: "epic", Sort: 51, Active: 1, Effect: map[string]interface{}{"color": "#8b5cf6"}},
	{Code: "pet_egg_slime", ItemType: "consumable", Name: "史莱姆蛋", Category: "pet_egg", Icon: "egg", Sort: 52, Effect: map[string]interface{}{"grantType": "pet", "petCode": "pet_slime"}},
	{Code: "pet_egg_cat", ItemType: "consumable", Name: "灵猫蛋", Category: "pet_egg", Icon: "egg", Sort: 53, Effect: map[string]interface{}{"grantType": "pet", "petCode": "pet_cat"}},
	{Code: "pet_egg_fox", ItemType: "consumable", Name: "灵狐蛋", Category: "pet_egg", Icon: "egg", Sort: 54, Effect: map[string]interface{}{"grantType": "pet", "petCode": "pet_fox"}},
	{Code: "pet_egg_owl", ItemType: "consumable", Name: "猫头鹰蛋", Category: "pet_egg", Icon: "egg", Sort: 55, Effect: map[string]interface{}{"grantType": "pet", "petCode": "pet_owl"}},
}

// PredefinedLotteryPool 奖池权重种子。
var PredefinedLotteryPool = []LotteryPoolSeed{
	{ItemCode: "exp_5", Probability: 0.25, Rarity: "common", Sort: 1, Active: 1},
	{ItemCode: "exp_10", Probability: 0.2, Rarity: "common", Sort: 2, Active: 1},
	{ItemCode: "ticket_1", Probability: 0.15, Rarity: "common", Sort: 3, Active: 1},
	{ItemCode: "exp_25", Probability: 0.12, Rarity: "rare", Sort: 5, Active: 1},
	{ItemCode: "buff_exp_small", Probability: 0.08, Rarity: "rare", Sort: 6, Active: 1},
	{ItemCode: "ticket_3", Probability: 0.05, Rarity: "rare", Sort: 7, Active: 1},
	{ItemCode: "exp_50", Probability: 0.06, Rarity: "epic", Sort: 10, Active: 1},
	{ItemCode: "buff_exp_medium", Probability: 0.04, Rarity: "epic", Sort: 11, Active: 1},
	{ItemCode: "buff_shield", Probability: 0.02, Rarity: "epic", Sort: 12, Active: 1},
	{ItemCode: "exp_100", Probability: 0.02, Rarity: "legendary", Sort: 15, Active: 1},
	{ItemCode: "ticket_10", Probability: 0.01, Rarity: "legendary", Sort: 16, Active: 1},
	{ItemCode: "lottery_title_writer", Probability: 0.03, Rarity: "rare", Sort: 20, Active: 1},
	{ItemCode: "lottery_frame_star", Probability: 0.015, Rarity: "epic", Sort: 21, Active: 1},
	{ItemCode: "pet_egg_slime", Probability: 0.01, Rarity: "epic", Sort: 22, Active: 1},
	{ItemCode: "pet_egg_cat", Probability: 0.008, Rarity: "epic", Sort: 23, Active: 1},
	{ItemCode: "pet_egg_fox", Probability: 0.006, Rarity: "epic", Sort: 24, Active: 1},
	{ItemCode: "pet_egg_owl", Probability: 0.004, Rarity: "legendary", Sort: 25, Active: 1},
}

// PredefinedQuests 预定义任务（对齐 Nest PREDEFINED_QUESTS 核心子集 + 完整结构）。
var PredefinedQuests = []QuestSeed{
	{Code: "daily_sign", Name: "每日签到", Description: "完成今日签到", Type: "daily", QuestSubtype: "daily", TargetAction: "sign_in", TargetCount: 1, ExpReward: 5, Sort: 1, Active: 1, Effect: map[string]interface{}{"ticketReward": 1}},
	{Code: "daily_comment", Name: "参与评论", Description: "发表2条评论", Type: "daily", QuestSubtype: "daily", TargetAction: "comment", TargetCount: 2, ExpReward: 10, Sort: 2, Active: 1},
	{Code: "daily_reply", Name: "积极回复", Description: "发表3条回复", Type: "daily", QuestSubtype: "daily", TargetAction: "reply", TargetCount: 3, ExpReward: 8, Sort: 3, Active: 1},
	{Code: "daily_article", Name: "发布文章", Description: "发布1篇文章", Type: "daily", QuestSubtype: "daily", TargetAction: "article", TargetCount: 1, ExpReward: 15, CurrencyReward: 5, Sort: 4, Active: 1},
	{Code: "daily_like", Name: "点赞文章", Description: "点赞3篇文章", Type: "daily", QuestSubtype: "daily", TargetAction: "like", TargetCount: 3, ExpReward: 8, Sort: 5, Active: 1},
	{Code: "daily_collect", Name: "收藏文章", Description: "收藏1篇文章", Type: "daily", QuestSubtype: "daily", TargetAction: "collect", TargetCount: 1, ExpReward: 5, CurrencyReward: 2, Sort: 6, Active: 1},
	{Code: "daily_msgboard", Name: "留言互动", Description: "在留言板发表1条留言", Type: "daily", QuestSubtype: "daily", TargetAction: "msgboard", TargetCount: 1, ExpReward: 5, Sort: 7, Active: 1},
	{Code: "daily_lottery", Name: "试试手气", Description: "完成1次抽奖", Type: "daily", QuestSubtype: "daily", TargetAction: "lottery_draw", TargetCount: 1, ExpReward: 5, Sort: 8, Active: 1, Effect: map[string]interface{}{"ticketReward": 1}},
	{Code: "daily_cheer", Name: "互助加油", Description: "为他人加油1次", Type: "daily", QuestSubtype: "daily", TargetAction: "social_cheer", TargetCount: 1, ExpReward: 5, Sort: 9, Active: 1},
	{Code: "bounty_comment_5", Name: "评论达人", Description: "今日发表5条评论", Type: "daily", QuestSubtype: "bounty", TargetAction: "comment", TargetCount: 5, ExpReward: 15, HpReward: 20, Sort: 20, Active: 1, Effect: map[string]interface{}{"ticketReward": 1}},
	{Code: "weekly_article_2", Name: "周更作者", Description: "本周发布2篇文章", Type: "weekly", QuestSubtype: "weekly", TargetAction: "article", TargetCount: 2, ExpReward: 30, Sort: 40, Active: 1},
	{Code: "weekly_sign_5", Name: "周签五连", Description: "本周签到5天", Type: "weekly", QuestSubtype: "weekly", TargetAction: "sign_in", TargetCount: 5, ExpReward: 20, Sort: 42, Active: 1},
	{Code: "special_first_tip", Name: "初次打赏", Description: "首次打赏一篇文章", Type: "special", QuestSubtype: "special", TargetAction: "tip", TargetCount: 1, ExpReward: 20, HpReward: 30, Sort: 30, Active: 1},
	{Code: "special_first_pet", Name: "驯兽入门", Description: "首次孵化或兑换宠物", Type: "special", QuestSubtype: "special", TargetAction: "pet_hatch", TargetCount: 1, ExpReward: 25, HpReward: 20, Sort: 31, Active: 1, Effect: map[string]interface{}{"ticketReward": 2}},
	{Code: "special_guild_join", Name: "公会新兵", Description: "加入一个公会", Type: "special", QuestSubtype: "special", TargetAction: "guild_join", TargetCount: 1, ExpReward: 30, HpReward: 25, Sort: 32, Active: 1},
}

// PredefinedAchievements 成就种子（核心子集，完整列表结构同 Nest）。
var PredefinedAchievements = []ItemConfigSeed{
	{Code: "first_article", ItemType: "achievement", Name: "初出茅庐", Description: "发布第一篇文章", Category: "creation", Icon: "pen", Rarity: "common", Sort: 1, Active: 1, Effect: map[string]interface{}{"maxProgress": 1, "expReward": 20, "trackEvent": "article", "currencyReward": 5, "hpReward": 5, "achievementConfigured": true}},
	{Code: "first_comment", ItemType: "achievement", Name: "初次发声", Description: "发表第一条评论", Category: "social", Icon: "chat", Rarity: "common", Sort: 10, Active: 1, Effect: map[string]interface{}{"maxProgress": 1, "expReward": 10, "trackEvent": "comment", "achievementConfigured": true}},
	{Code: "sign_7days", ItemType: "achievement", Name: "七日之约", Description: "累计签到7天", Category: "sign", Icon: "calendar", Rarity: "rare", Sort: 20, Active: 1, Effect: map[string]interface{}{"maxProgress": 7, "expReward": 30, "trackEvent": "sign_in", "ticketReward": 1, "achievementConfigured": true}},
	{Code: "first_lottery", ItemType: "achievement", Name: "初试手气", Description: "首次抽奖", Category: "lottery", Icon: "dice", Rarity: "common", Sort: 70, Active: 1, Effect: map[string]interface{}{"maxProgress": 1, "expReward": 10, "trackEvent": "lottery_draw", "ticketReward": 1, "achievementConfigured": true}},
	{Code: "level_2", ItemType: "achievement", Name: "新手冒险者", Description: "等级达到LV2", Category: "special", Icon: "star", Rarity: "common", Sort: 39, Active: 1, Effect: map[string]interface{}{"maxProgress": 2, "expReward": 20, "trackEvent": "level_up", "currencyReward": 20, "items": []string{"frame_a"}, "achievementConfigured": true}},
}

// PredefinedActivities 活动种子。
var PredefinedActivities = []ActivitySeed{
	{Code: "season_spring_2026", Name: "2026 春季赛季", Description: "春季冒险赛季，经验获取提升", ActivityType: "season", StartTime: "2026-01-01T00:00:00+08:00", EndTime: "2026-12-31T23:59:59+08:00", ExpBuffRate: 1.2, PosterURL: "/images/rpg/season-spring-2026-poster.png", Active: 1},
	{Code: "festival_spring_2026", Name: "2026 春节", Description: "春节期间经验额外提升", ActivityType: "festival", StartTime: "2026-01-20T00:00:00+08:00", EndTime: "2026-02-10T23:59:59+08:00", ExpBuffRate: 1.15, Active: 1},
}

// UpsertItemConfigSeeds 幂等 upsert 物品配置种子。
func UpsertItemConfigSeeds(ctx context.Context, repo *rpgrepo.RpgRepo, seeds []ItemConfigSeed, log *zap.Logger) error {
	for _, def := range seeds {
		if def.Code == "" {
			continue
		}
		effect := def.Effect
		if effect == nil {
			effect = map[string]interface{}{}
		}
		effectBytes, _ := json.Marshal(effect)
		effectStr := string(effectBytes)
		active := def.Active
		if active == 0 {
			active = 1
		}
		sort := def.Sort
		if sort == 0 {
			sort = 10
		}
		exists, err := repo.FindItemConfigByCode(ctx, def.Code)
		if err != nil && !ent.IsNotFound(err) {
			return err
		}
		if exists == nil {
			_, err = repo.CreateItemConfig(ctx, &ent.RpgItemConfig{
				Code:        def.Code,
				ItemType:    def.ItemType,
				Name:        def.Name,
				Description: def.Description,
				Category:    def.Category,
				Icon:        def.Icon,
				Rarity:      def.Rarity,
				Sort:        sort,
				Active:      active,
				IsHidden:    def.IsHidden,
				EffectJson:  &effectStr,
			})
		} else {
			err = repo.UpdateItemConfig(ctx, exists.ID, map[string]interface{}{
				rpgitemconfig.FieldName:        def.Name,
				rpgitemconfig.FieldDescription: def.Description,
				rpgitemconfig.FieldItemType:    def.ItemType,
				rpgitemconfig.FieldCategory:    def.Category,
				rpgitemconfig.FieldIcon:        def.Icon,
				rpgitemconfig.FieldRarity:      def.Rarity,
				rpgitemconfig.FieldSort:        sort,
				rpgitemconfig.FieldActive:      active,
				rpgitemconfig.FieldEffectJson:  effectStr,
			})
		}
		if err != nil {
			return err
		}
	}
	if log != nil {
		log.Info("物品配置种子同步完成", zap.Int("count", len(seeds)))
	}
	return nil
}

// UpsertQuestSeeds 幂等 upsert 任务种子。
func UpsertQuestSeeds(ctx context.Context, repo *rpgrepo.RpgRepo, seeds []QuestSeed, log *zap.Logger) error {
	for _, def := range seeds {
		effectStr := ""
		if def.Effect != nil {
			b, _ := json.Marshal(def.Effect)
			effectStr = string(b)
		}
		active := def.Active
		if active == 0 {
			active = 1
		}
		exists, err := repo.FindQuestByCode(ctx, def.Code)
		if err != nil && !ent.IsNotFound(err) {
			return err
		}
		if exists == nil {
			var effect *string
			if effectStr != "" {
				effect = &effectStr
			}
			_, err = repo.CreateQuest(ctx, &ent.RpgQuest{
				Code:           def.Code,
				Name:           def.Name,
				Description:    def.Description,
				Type:           def.Type,
				QuestSubtype:   def.QuestSubtype,
				TargetAction:   def.TargetAction,
				TargetCount:    def.TargetCount,
				ExpReward:      def.ExpReward,
				HpReward:       def.HpReward,
				CurrencyReward: def.CurrencyReward,
				Sort:           def.Sort,
				Active:         active,
				EffectJson:     effect,
			})
		} else {
			patch := map[string]interface{}{
				rpgquest.FieldName:           def.Name,
				rpgquest.FieldDescription:    def.Description,
				rpgquest.FieldTargetCount:    def.TargetCount,
				rpgquest.FieldExpReward:      def.ExpReward,
				rpgquest.FieldHpReward:       def.HpReward,
				rpgquest.FieldCurrencyReward: def.CurrencyReward,
				rpgquest.FieldSort:           def.Sort,
				rpgquest.FieldActive:         active,
			}
			if effectStr != "" {
				patch[rpgquest.FieldEffectJson] = effectStr
			}
			err = repo.UpdateQuest(ctx, exists.ID, patch)
		}
		if err != nil {
			return err
		}
	}
	if log != nil {
		log.Info("任务种子同步完成", zap.Int("count", len(seeds)))
	}
	return nil
}

// UpsertLotteryPoolSeeds 幂等 upsert 抽奖池。
func UpsertLotteryPoolSeeds(ctx context.Context, repo *rpgrepo.RpgRepo, seeds []LotteryPoolSeed, log *zap.Logger) error {
	for _, def := range seeds {
		active := def.Active
		if active == 0 {
			active = 1
		}
		exists, err := repo.FindLotteryPoolByItemCode(ctx, def.ItemCode)
		if err != nil && !ent.IsNotFound(err) {
			return err
		}
		if exists == nil {
			_, err = repo.CreateLotteryPool(ctx, &ent.RpgLotteryPool{
				ItemCode:    def.ItemCode,
				Probability: def.Probability,
				Rarity:      def.Rarity,
				Sort:        def.Sort,
				Active:      active,
			})
		} else {
			err = repo.UpdateLotteryPool(ctx, exists.ID, map[string]interface{}{
				rpglotterypool.FieldProbability: def.Probability,
				rpglotterypool.FieldRarity:      def.Rarity,
				rpglotterypool.FieldSort:        def.Sort,
				rpglotterypool.FieldActive:      active,
			})
		}
		if err != nil {
			return err
		}
	}
	if log != nil {
		log.Info("抽奖池种子同步完成", zap.Int("count", len(seeds)))
	}
	return nil
}

// UpsertActivitySeeds 幂等 insert 缺失活动（不覆盖已有行）。
func UpsertActivitySeeds(ctx context.Context, repo *rpgrepo.RpgRepo, seeds []ActivitySeed, log *zap.Logger) error {
	for _, def := range seeds {
		exists, err := repo.FindActivityByCode(ctx, def.Code)
		if err != nil && !ent.IsNotFound(err) {
			return err
		}
		if exists != nil {
			continue
		}
		start, _ := parseTime(def.StartTime)
		end, _ := parseTime(def.EndTime)
		active := def.Active
		if active == 0 {
			active = 1
		}
		_, err = repo.CreateActivity(ctx, &ent.RpgActivity{
			Code:         def.Code,
			Name:         def.Name,
			Description:  def.Description,
			ActivityType: def.ActivityType,
			StartTime:    start,
			EndTime:      end,
			ExpBuffRate:  def.ExpBuffRate,
			PosterUrl:    def.PosterURL,
			Active:       active,
		})
		if err != nil {
			return err
		}
	}
	if log != nil {
		log.Info("活动种子同步完成", zap.Int("count", len(seeds)))
	}
	return nil
}

// SyncAllPredefined 启动时同步全部预定义种子。
func SyncAllPredefined(ctx context.Context, repo *rpgrepo.RpgRepo, log *zap.Logger) error {
	allItems := append([]ItemConfigSeed{CurrencyItemDef}, PredefinedCosmeticSeeds...)
	allItems = append(allItems, PredefinedLotteryItems...)
	allItems = append(allItems, PredefinedAchievements...)
	if err := UpsertItemConfigSeeds(ctx, repo, allItems, log); err != nil {
		return err
	}
	if err := UpsertQuestSeeds(ctx, repo, PredefinedQuests, log); err != nil {
		return err
	}
	if err := UpsertLotteryPoolSeeds(ctx, repo, PredefinedLotteryPool, log); err != nil {
		return err
	}
	return UpsertActivitySeeds(ctx, repo, PredefinedActivities, log)
}

func parseTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}
