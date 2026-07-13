// Package activity 赛季/节日活动与经验 Buff。
package activity

import (
	"context"
	"time"

	"github.com/Jiang-Xia/blog-server-go/services/rpg/internal/event"
	rpgrepo "github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/repo"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/seeds"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/util"
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

// GetCurrentActivitiesOverview C 端活动概览（与 Nest activity.service 对齐）。
func (s *Service) GetCurrentActivitiesOverview(ctx context.Context) (interface{}, error) {
	activities, err := s.GetCurrentActivities(ctx)
	if err != nil {
		return nil, err
	}
	if len(activities) == 0 {
		return nil, nil
	}
	var season *ent.RpgActivity
	limitedTime := make([]*ent.RpgActivity, 0, len(activities))
	maxRate := 1.0
	for _, a := range activities {
		if a.ExpBuffRate > maxRate {
			maxRate = a.ExpBuffRate
		}
		if a.ActivityType == "season" {
			season = a
		} else {
			limitedTime = append(limitedTime, a)
		}
	}
	if season == nil && len(limitedTime) == 0 {
		return nil, nil
	}
	limitedOut := make([]map[string]interface{}, 0, len(limitedTime))
	for _, a := range limitedTime {
		limitedOut = append(limitedOut, util.FormatActivitySummary(a))
	}
	result := map[string]interface{}{
		"season":               util.FormatActivitySummary(season),
		"limitedTime":          limitedOut,
		"effectiveExpBuffRate": util.RoundExpBuffRate(maxRate),
	}
	return result, nil
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
