// Package level 等级/经验与签到业务。
package level

import (
	"context"
	"fmt"
	"time"

	"github.com/Jiang-Xia/blog-server-go/pkg/redisutil"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/ent"
	rpgcore "github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/core"
	rpgnotify "github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/notify"
	rpgrepo "github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/repo"
)

// LevelAchievementTracker 等级成就追踪（避免 level ↔ achievement 循环依赖）。
type LevelAchievementTracker interface {
	TrackLevelUp(ctx context.Context, uid, newLevel int) error
}

// LevelService 等级阈值计算、经验发放与升级判定。
type LevelService struct {
	rpg         *rpgcore.RpgService
	repo        *rpgrepo.RpgRepo
	notify      *rpgnotify.RpgNotifyService
	redis       *redisutil.Store
	achievement LevelAchievementTracker
}

// NewLevelService 构造 LevelService。
func NewLevelService(
	rpg *rpgcore.RpgService,
	repo *rpgrepo.RpgRepo,
	notify *rpgnotify.RpgNotifyService,
	redis *redisutil.Store,
) *LevelService {
	return &LevelService{rpg: rpg, repo: repo, notify: notify, redis: redis}
}

// SetAchievementTracker 延迟注入等级成就追踪。
func (s *LevelService) SetAchievementTracker(t LevelAchievementTracker) {
	s.achievement = t
}

// GetLevelThreshold 计算指定等级所需最低经验；公式 level * (level-1) * 50，LV1=0。
func (s *LevelService) GetLevelThreshold(level int) int {
	if level <= 1 {
		return 0
	}
	return level * (level - 1) * 50
}

// GetNextLevelThreshold 获取当前等级下一级所需经验值。
func (s *LevelService) GetNextLevelThreshold(currentLevel int) int {
	return s.GetLevelThreshold(currentLevel + 1)
}

// GetExpProgress 获取当前经验在当前等级中的进度。
func (s *LevelService) GetExpProgress(rpg *ent.Rpg) rpgcore.ExpProgress {
	currentThreshold := s.GetLevelThreshold(rpg.Level)
	nextThreshold := s.GetNextLevelThreshold(rpg.Level)
	current := rpg.Exp - currentThreshold
	required := nextThreshold - currentThreshold
	percent := 100
	if required > 0 {
		percent = min(100, (current*100)/required)
	}
	return rpgcore.ExpProgress{Current: current, Required: required, Percent: percent}
}

// AddExp 增加经验并处理升级；dailyLimit 为每日上限，0 表示无限制。
func (s *LevelService) AddExp(ctx context.Context, uid, amount int, reason rpgcore.ExpReason, dailyLimit int) (*rpgcore.LevelUpResult, error) {
	if dailyLimit > 0 && s.redis != nil {
		limited, err := s.applyDailyLimit(ctx, uid, reason, dailyLimit, amount)
		if err != nil {
			return nil, err
		}
		if limited <= 0 {
			return nil, nil
		}
		amount = limited
	}

	rpg, err := s.rpg.GetOrCreateRpg(ctx, uid)
	if err != nil {
		return nil, err
	}
	rpg.Exp += amount

	levelUpPartial := s.checkAndLevelUp(rpg)
	if _, err := s.rpg.SaveRpg(ctx, rpg); err != nil {
		return nil, err
	}

	if levelUpPartial == nil {
		if amount > 0 && s.notify != nil {
			s.notify.NotifyExpGain(ctx, uid, amount, string(reason))
		}
		return nil, nil
	}

	unlocked, err := s.getLevelRewardsUnlockedInRange(ctx, levelUpPartial.OldLevel, levelUpPartial.NewLevel)
	if err != nil {
		return nil, err
	}
	result := &rpgcore.LevelUpResult{
		OldLevel:        levelUpPartial.OldLevel,
		NewLevel:        levelUpPartial.NewLevel,
		UnlockedRewards: unlocked,
	}
	if s.notify != nil {
		s.notify.NotifyLevelUp(ctx, uid, result)
	}
	if s.achievement != nil {
		_ = s.achievement.TrackLevelUp(ctx, uid, levelUpPartial.NewLevel)
	}
	return result, nil
}

func (s *LevelService) applyDailyLimit(ctx context.Context, uid int, reason rpgcore.ExpReason, dailyLimit, amount int) (int, error) {
	today := time.Now().Format("2006-01-02")
	cacheKey := fmt.Sprintf("rpg:exp:%d:%s:%s", uid, reason, today)
	currentExp := int(redisutil.ParseInt(mustGet(s.redis.Get(ctx, cacheKey))))
	remaining := dailyLimit - currentExp
	if remaining <= 0 {
		return 0, nil
	}
	if amount > remaining {
		amount = remaining
	}
	now := time.Now()
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
	ttlSec := int(endOfDay.Sub(now).Seconds()) + 60
	if err := s.redis.Set(ctx, cacheKey, fmt.Sprintf("%d", currentExp+amount), ttlSec); err != nil {
		return 0, err
	}
	return amount, nil
}

func mustGet(v string, err error) string {
	if err != nil {
		return ""
	}
	return v
}

func (s *LevelService) checkAndLevelUp(rpg *ent.Rpg) *struct {
	OldLevel int
	NewLevel int
} {
	newLevel := rpg.Level
	for rpg.Exp >= s.GetLevelThreshold(newLevel+1) {
		newLevel++
	}
	if newLevel == rpg.Level {
		return nil
	}
	oldLevel := rpg.Level
	rpg.Level = newLevel
	return &struct {
		OldLevel int
		NewLevel int
	}{OldLevel: oldLevel, NewLevel: newLevel}
}

// GetLevelRewards 获取某等级及以下可解锁的全部奖励（前端展示）。
func (s *LevelService) GetLevelRewards(ctx context.Context, level int) ([]rpgcore.LevelReward, error) {
	all, err := s.getAllLevelRewards(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]rpgcore.LevelReward, 0)
	for _, r := range all {
		if r.Level <= level {
			out = append(out, r)
		}
	}
	return out, nil
}

func (s *LevelService) getAllLevelRewards(ctx context.Context) ([]rpgcore.LevelReward, error) {
	rows, err := s.repo.ListActiveLevelRewards(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]rpgcore.LevelReward, 0, len(rows))
	for _, row := range rows {
		out = append(out, rpgcore.LevelReward{
			Level:          row.Level,
			CurrencyReward: row.CurrencyReward,
			AvatarFrame:    row.AvatarFrame,
			Title:          row.Title,
		})
	}
	return out, nil
}

func (s *LevelService) getLevelRewardsUnlockedInRange(ctx context.Context, oldLevel, newLevel int) ([]rpgcore.LevelReward, error) {
	all, err := s.getAllLevelRewards(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]rpgcore.LevelReward, 0)
	for _, r := range all {
		if r.Level > oldLevel && r.Level <= newLevel {
			out = append(out, r)
		}
	}
	if out == nil {
		out = []rpgcore.LevelReward{}
	}
	return out, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
