// Package achievement 成就进度追踪与奖励发放。
package achievement

import (
	"context"
	"time"

	rpgconst "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/constants"
	rpgcore "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/core"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/inventory"
	rpglevel "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/level"
	rpgnotify "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/notify"
	rpgrepo "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/repo"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/seeds"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/util"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
	"go.uber.org/zap"
)

// Service 成就业务。
type Service struct {
	repo      *rpgrepo.RpgRepo
	level     *rpglevel.LevelService
	inventory *inventory.Service
	core      *rpgcore.RpgService
	notify    *rpgnotify.RpgNotifyService
	log       *zap.Logger
}

// NewService 构造成就 Service。
func NewService(
	repo *rpgrepo.RpgRepo,
	level *rpglevel.LevelService,
	inventory *inventory.Service,
	core *rpgcore.RpgService,
	log *zap.Logger,
) *Service {
	return &Service{repo: repo, level: level, inventory: inventory, core: core, log: log}
}

// SetNotify 延迟注入 WS 推送。
func (s *Service) SetNotify(n *rpgnotify.RpgNotifyService) {
	s.notify = n
}

// SyncPredefinedAchievements 启动时同步成就种子。
func (s *Service) SyncPredefinedAchievements(ctx context.Context) error {
	return seeds.UpsertItemConfigSeeds(ctx, s.repo, seeds.PredefinedAchievements, s.log)
}

// GetMyAchievements 获取用户成就列表（含进度与稀有度展示字段）。
func (s *Service) GetMyAchievements(ctx context.Context, uid int) ([]map[string]interface{}, error) {
	configs, err := s.repo.ListAchievementConfigs(ctx)
	if err != nil {
		return nil, err
	}
	progressRows, err := s.repo.ListAchievementsByUID(ctx, uid)
	if err != nil {
		return nil, err
	}
	progMap := map[string]*ent.RpgUserAchievement{}
	for _, p := range progressRows {
		progMap[p.AchievementCode] = p
	}
	out := make([]map[string]interface{}, 0, len(configs))
	for _, cfg := range configs {
		effect := util.ParseEffectJSON(cfg.EffectJson)
		if v, ok := effect["achievementConfigured"].(bool); ok && !v {
			continue
		}
		maxProgress := intFromEffect(effect["maxProgress"])
		rd := rpgconst.GetRarityDisplay(cfg.Rarity)
		item := map[string]interface{}{
			"id":          cfg.ID,
			"code":        cfg.Code,
			"name":        cfg.Name,
			"description": cfg.Description,
			"category":    cfg.Category,
			"icon":        cfg.Icon,
			"rarity":      cfg.Rarity,
			"rarityLabel": rd.Label,
			"rarityColor": rd.Color,
			"rarityIcon":  rd.Icon,
			"maxProgress": maxProgress,
			"expReward":   intFromEffect(effect["expReward"]),
			"sort":        cfg.Sort,
			"badge":       map[string]interface{}{"color": rd.Color},
			"progress":    0,
			"completed":   false,
			"completedAt": nil,
		}
		if p, ok := progMap[cfg.Code]; ok {
			item["progress"] = p.Progress
			item["completed"] = p.Completed == 1
		}
		out = append(out, item)
	}
	return out, nil
}

// TrackProgress 按 trackEvent 递增成就进度。
func (s *Service) TrackProgress(ctx context.Context, uid int, event string) error {
	configs, err := s.repo.ListAchievementConfigs(ctx)
	if err != nil {
		return err
	}
	for _, cfg := range configs {
		effect := util.ParseEffectJSON(cfg.EffectJson)
		if trackEvent, _ := effect["trackEvent"].(string); trackEvent != event {
			continue
		}
		if err := s.incrementAchievement(ctx, uid, cfg, effect, 1); err != nil {
			return err
		}
	}
	return nil
}

// TrackSignIn 签到类成就（含 consecutive_sign）。
func (s *Service) TrackSignIn(ctx context.Context, uid int, totalDays, consecutiveDays int) error {
	_ = s.trackAbsolute(ctx, uid, "sign_in", totalDays)
	return s.trackAbsolute(ctx, uid, "consecutive_sign", consecutiveDays)
}

// TrackLevelUp 等级成就。
func (s *Service) TrackLevelUp(ctx context.Context, uid, newLevel int) error {
	return s.trackAbsolute(ctx, uid, "level_up", newLevel)
}

// GetAllLevelRewardCatalog 等级奖励一览（对齐 Nest：trackEvent=level_up 的成就配置）。
func (s *Service) GetAllLevelRewardCatalog(ctx context.Context) ([]rpgcore.LevelReward, error) {
	configs, err := s.listLevelUpAchievementConfigs(ctx)
	if err != nil {
		return nil, err
	}
	return s.buildLevelRewardCatalog(ctx, configs)
}

func (s *Service) listLevelUpAchievementConfigs(ctx context.Context) ([]*ent.RpgItemConfig, error) {
	configs, err := s.repo.ListAchievementConfigs(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*ent.RpgItemConfig, 0)
	for _, cfg := range configs {
		effect := util.ParseEffectJSON(cfg.EffectJson)
		if v, ok := effect["achievementConfigured"].(bool); ok && !v {
			continue
		}
		if trackEvent, _ := effect["trackEvent"].(string); trackEvent != "level_up" {
			continue
		}
		out = append(out, cfg)
	}
	return out, nil
}

func (s *Service) buildLevelRewardCatalog(ctx context.Context, configs []*ent.RpgItemConfig) ([]rpgcore.LevelReward, error) {
	itemCodes := map[string]struct{}{}
	for _, cfg := range configs {
		effect := util.ParseEffectJSON(cfg.EffectJson)
		for _, code := range effectStringSlice(effect["items"]) {
			itemCodes[code] = struct{}{}
		}
	}
	typeByCode := map[string]string{}
	if len(itemCodes) > 0 {
		codes := make([]string, 0, len(itemCodes))
		for code := range itemCodes {
			codes = append(codes, code)
		}
		rows, err := s.repo.ListItemConfigsByCodes(ctx, codes)
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			typeByCode[row.Code] = row.ItemType
		}
	}

	out := make([]rpgcore.LevelReward, 0, len(configs))
	for _, cfg := range configs {
		effect := util.ParseEffectJSON(cfg.EffectJson)
		var avatarFrame, title string
		for _, code := range effectStringSlice(effect["items"]) {
			switch typeByCode[code] {
			case "avatar_frame":
				avatarFrame = code
			case "title":
				title = code
			}
		}
		out = append(out, rpgcore.LevelReward{
			Level:          intFromEffect(effect["maxProgress"]),
			CurrencyReward: intFromEffect(effect["currencyReward"]),
			AvatarFrame:    avatarFrame,
			Title:          title,
		})
	}
	// 按等级升序（与 Nest sort 一致）
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j].Level < out[i].Level {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	if out == nil {
		out = []rpgcore.LevelReward{}
	}
	return out, nil
}

func effectStringSlice(v interface{}) []string {
	switch t := v.(type) {
	case []interface{}:
		out := make([]string, 0, len(t))
		for _, item := range t {
			if s, ok := item.(string); ok && s != "" {
				out = append(out, s)
			}
		}
		return out
	case []string:
		return t
	default:
		return nil
	}
}

// TrackReputation 声望阈值成就。
func (s *Service) TrackReputation(ctx context.Context, uid, reputation int) error {
	return s.trackAbsolute(ctx, uid, "reputation", reputation)
}

func (s *Service) trackAbsolute(ctx context.Context, uid int, event string, absolute int) error {
	configs, err := s.repo.ListAchievementConfigs(ctx)
	if err != nil {
		return err
	}
	for _, cfg := range configs {
		effect := util.ParseEffectJSON(cfg.EffectJson)
		if trackEvent, _ := effect["trackEvent"].(string); trackEvent != event {
			continue
		}
		maxProgress := intFromEffect(effect["maxProgress"])
		if absolute < maxProgress {
			continue
		}
		if err := s.completeAchievement(ctx, uid, cfg, effect, maxProgress); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) incrementAchievement(ctx context.Context, uid int, cfg *ent.RpgItemConfig, effect map[string]interface{}, delta int) error {
	maxProgress := intFromEffect(effect["maxProgress"])
	if maxProgress <= 0 {
		maxProgress = 1
	}
	prog, err := s.repo.FindAchievementProgress(ctx, uid, cfg.Code)
	if ent.IsNotFound(err) {
		prog = &ent.RpgUserAchievement{UID: uid, AchievementCode: cfg.Code}
	} else if err != nil {
		return err
	}
	if prog.Completed == 1 {
		return nil
	}
	prog.Progress += delta
	if prog.Progress >= maxProgress {
		return s.completeAchievement(ctx, uid, cfg, effect, maxProgress)
	}
	_, err = s.repo.SaveAchievementProgress(ctx, prog)
	return err
}

func (s *Service) completeAchievement(ctx context.Context, uid int, cfg *ent.RpgItemConfig, effect map[string]interface{}, progress int) error {
	now := time.Now()
	prog, err := s.repo.FindAchievementProgress(ctx, uid, cfg.Code)
	if ent.IsNotFound(err) {
		prog = &ent.RpgUserAchievement{UID: uid, AchievementCode: cfg.Code}
	} else if err != nil {
		return err
	}
	if prog.Completed == 1 {
		return nil
	}
	prog.Progress = progress
	prog.Completed = 1
	prog.CompletedAt = &now
	if _, err := s.repo.SaveAchievementProgress(ctx, prog); err != nil {
		return err
	}
	if exp := intFromEffect(effect["expReward"]); exp > 0 {
		_, _ = s.level.AddExp(ctx, uid, exp, rpgcore.ExpReasonAchievement, 0)
	}
	if cur := intFromEffect(effect["currencyReward"]); cur > 0 {
		_, _ = s.inventory.AdjustCurrency(ctx, uid, cur, "achievement")
	}
	if items, ok := effect["items"].([]interface{}); ok {
		for _, it := range items {
			if code, ok := it.(string); ok && code != "" {
				_ = s.inventory.GrantItem(ctx, uid, code, "achievement", 1)
			}
		}
	}
	if s.notify != nil {
		s.notify.NotifyAchievementComplete(ctx, uid, rpgnotify.BuildAchievementCompletePayload(cfg, effect))
	}
	return nil
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
