// Package social 声望累计与发文加成。
package social

import (
	"context"

	rpgcore "github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/core"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/constants"
)

// ReputationService 声望业务。
type ReputationService struct {
	core        *rpgcore.RpgService
	achievement AchievementTracker
}

// NewReputationService 构造 ReputationService。
func NewReputationService(core *rpgcore.RpgService, achievement AchievementTracker) *ReputationService {
	return &ReputationService{core: core, achievement: achievement}
}

// AddReputation 累加用户声望。
func (s *ReputationService) AddReputation(ctx context.Context, uid, amount int, reason string) (int, error) {
	if amount <= 0 {
		return 0, nil
	}
	rpg, err := s.core.GetOrCreateRpg(ctx, uid)
	if err != nil {
		return 0, err
	}
	rpg.Reputation += amount
	saved, err := s.core.SaveRpg(ctx, rpg)
	if err != nil {
		return 0, err
	}
	if s.achievement != nil {
		_ = s.achievement.TrackReputation(ctx, uid, saved.Reputation)
	}
	_ = reason
	return saved.Reputation, nil
}

// GetReputation 查询声望。
func (s *ReputationService) GetReputation(ctx context.Context, uid int) (int, error) {
	rpg, err := s.core.FindByUid(ctx, uid)
	if err != nil || rpg == nil {
		return 0, err
	}
	return rpg.Reputation, nil
}

// GetPublishExpBoostRate 声望发文加成倍率（每100声望+5%，上限50%）。
func (s *ReputationService) GetPublishExpBoostRate(reputation int) float64 {
	bonus := float64(reputation/100) * 0.05
	if bonus > 0.5 {
		bonus = 0.5
	}
	return 1 + bonus
}

// GetReputationForAction 互动声望映射。
func (s *ReputationService) GetReputationForAction(action string) int {
	switch action {
	case "like":
		return constants.Economy.ReputationLike
	case "comment":
		return constants.Economy.ReputationComment
	case "collect":
		return constants.Economy.ReputationCollect
	case "view":
		return constants.Economy.ArticleViewReputation
	case "flower":
		return constants.Economy.FlowerReputation
	default:
		return 0
	}
}
