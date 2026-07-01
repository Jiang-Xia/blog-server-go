// Package activity 赛季/节日活动与经验 Buff。
package activity

import (
	"context"
	"time"

	"github.com/Jiang-Xia/blog-server-go/services/rpg/internal/event"
	rpgrepo "github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/repo"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/seeds"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/ent"
	"go.uber.org/zap"
)

// Service 活动业务。
type Service struct {
	repo      *rpgrepo.RpgRepo
	publisher *event.Publisher
	log       *zap.Logger
}

// NewService 构造活动 Service。
func NewService(repo *rpgrepo.RpgRepo, publisher *event.Publisher, log *zap.Logger) *Service {
	return &Service{repo: repo, publisher: publisher, log: log}
}

// SyncPredefinedActivities 启动时同步活动种子。
func (s *Service) SyncPredefinedActivities(ctx context.Context) error {
	return seeds.UpsertActivitySeeds(ctx, s.repo, seeds.PredefinedActivities, s.log)
}

// GetCurrentActivities 当前有效活动列表。
func (s *Service) GetCurrentActivities(ctx context.Context) ([]*ent.RpgActivity, error) {
	return s.repo.ListCurrentActivities(ctx, time.Now())
}

// GetExpBuffRate 多活动并存时取最高 expBuffRate。
func (s *Service) GetExpBuffRate(ctx context.Context) (float64, error) {
	activities, err := s.GetCurrentActivities(ctx)
	if err != nil {
		return 1, err
	}
	if len(activities) == 0 {
		return 1, nil
	}
	maxRate := 1.0
	for _, a := range activities {
		if a.ExpBuffRate > maxRate {
			maxRate = a.ExpBuffRate
		}
	}
	return maxRate, nil
}

// GetCurrentSeasonKey 当前赛季 Redis 键。
func (s *Service) GetCurrentSeasonKey(ctx context.Context) (string, error) {
	activities, err := s.GetCurrentActivities(ctx)
	if err != nil {
		return "", err
	}
	for _, a := range activities {
		if a.ActivityType == "season" {
			return a.Code, nil
		}
	}
	return "season_" + time.Now().Format("2006"), nil
}

// SharePoster 分享赛季海报（发布领域事件）。
func (s *Service) SharePoster(ctx context.Context, uid int, activityCode string) {
	if s.publisher == nil {
		return
	}
	s.publisher.Publish(ctx, event.EventSeasonPosterShared, map[string]interface{}{
		"uid":          uid,
		"activityCode": activityCode,
	})
}
