// Package social 用户间互动（加油/砸蛋/送花）。
package social

import (
	"context"
	"fmt"
	"time"

	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/pkg/redisutil"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/constants"
	rpgcore "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/core"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/inventory"
	rpgrepo "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/repo"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/util"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
)

// InteractService 社交互动业务。
type InteractService struct {
	repo        *rpgrepo.RpgRepo
	core        *rpgcore.RpgService
	inventory   *inventory.Service
	reputation  *ReputationService
	redis       *redisutil.Store
	achievement AchievementTracker
	quest       QuestTracker
}

// AchievementTracker 成就追踪。
type AchievementTracker interface {
	TrackProgress(ctx context.Context, uid int, event string) error
	TrackReputation(ctx context.Context, uid, reputation int) error
}

// QuestTracker 任务追踪。
type QuestTracker interface {
	TrackProgress(ctx context.Context, uid int, action string) error
}

// NewInteractService 构造 InteractService。
func NewInteractService(
	repo *rpgrepo.RpgRepo,
	core *rpgcore.RpgService,
	inventory *inventory.Service,
	reputation *ReputationService,
	redis *redisutil.Store,
	achievement AchievementTracker,
	quest QuestTracker,
) *InteractService {
	return &InteractService{
		repo:        repo,
		core:        core,
		inventory:   inventory,
		reputation:  reputation,
		redis:       redis,
		achievement: achievement,
		quest:       quest,
	}
}

// Cheer 为他人加油。
func (s *InteractService) Cheer(ctx context.Context, fromUID, toUID int) (map[string]interface{}, error) {
	return s.interact(ctx, fromUID, toUID, constants.SocialActionCheer)
}

// Egg 砸蛋。
func (s *InteractService) Egg(ctx context.Context, fromUID, toUID int) (map[string]interface{}, error) {
	return s.interact(ctx, fromUID, toUID, constants.SocialActionEgg)
}

// Flower 送花。
func (s *InteractService) Flower(ctx context.Context, fromUID, toUID int) (map[string]interface{}, error) {
	return s.interact(ctx, fromUID, toUID, constants.SocialActionFlower)
}

func (s *InteractService) interact(ctx context.Context, fromUID, toUID int, action constants.SocialAction) (map[string]interface{}, error) {
	if fromUID == toUID {
		return nil, errcode.WithMessage(errcode.InvalidParam, "不能对自己操作")
	}
	target, err := s.core.GetOrCreateRpg(ctx, toUID)
	if err != nil {
		return nil, err
	}
	limit, cost, hpDelta, repDelta := actionParams(action)
	if err := s.checkDailyLimit(ctx, fromUID, action, limit); err != nil {
		return nil, err
	}
	if cost > 0 {
		if _, err := s.inventory.AdjustCurrency(ctx, fromUID, -cost, string(action)); err != nil {
			return nil, err
		}
	}
	if hpDelta != 0 {
		target.LifeValue = clampLife(target.LifeValue + hpDelta)
		if _, err := s.core.SaveRpg(ctx, target); err != nil {
			return nil, err
		}
	}
	if repDelta > 0 {
		_, _ = s.reputation.AddReputation(ctx, toUID, repDelta, string(action))
	}
	_, _ = s.repo.CreateSocialLog(ctx, &ent.RpgUserSocialLog{
		FromUid:      fromUID,
		ToUid:        toUID,
		Action:       string(action),
		CostCurrency: cost,
		HpDelta:      hpDelta,
	})
	trackAction := "social_" + string(action)
	if s.achievement != nil {
		_ = s.achievement.TrackProgress(ctx, fromUID, trackAction)
	}
	if s.quest != nil {
		_ = s.quest.TrackProgress(ctx, fromUID, trackAction)
	}
	return map[string]interface{}{
		"success":      true,
		"action":       action,
		"hpDelta":      hpDelta,
		"costCurrency": cost,
	}, nil
}

func actionParams(action constants.SocialAction) (limit, cost, hpDelta, repDelta int) {
	switch action {
	case constants.SocialActionCheer:
		return constants.Economy.CheerDailyLimit, 0, constants.Economy.CheerHP, 0
	case constants.SocialActionEgg:
		return constants.Economy.EggDailyLimit, constants.Economy.EggCost, constants.Economy.EggHP, 0
	case constants.SocialActionFlower:
		return constants.Economy.FlowerDailyLimit, constants.Economy.FlowerCost, 0, constants.Economy.FlowerReputation
	default:
		return 0, 0, 0, 0
	}
}

func (s *InteractService) checkDailyLimit(ctx context.Context, uid int, action constants.SocialAction, limit int) error {
	if s.redis == nil || limit <= 0 {
		return nil
	}
	key := fmt.Sprintf("rpg:social:%d:%s:%s", uid, action, time.Now().Format("2006-01-02"))
	count := int(redisutil.ParseInt(mustGet(s.redis.Get(ctx, key))))
	if count >= limit {
		return errcode.WithMessage(errcode.TooManyRequests, "今日次数已达上限")
	}
	return s.redis.Set(ctx, key, fmt.Sprintf("%d", count+1), util.SecondsUntilMidnight())
}

func clampLife(v int) int {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}

func mustGet(v string, err error) string {
	if err != nil {
		return ""
	}
	return v
}
