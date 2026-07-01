// Package activity 活动相关定时任务。
package activity

import (
	"context"
	"time"

	"github.com/Jiang-Xia/blog-server-go/services/rpg/ent"
	rpgnotify "github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/notify"
	rpgrepo "github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/repo"
	"go.uber.org/zap"
)

// NotifyScheduler 每日 8:00 检查活动开始/结束并 WS 推送。
type NotifyScheduler struct {
	repo   *rpgrepo.RpgRepo
	notify *rpgnotify.RpgNotifyService
	log    *zap.Logger
}

// NewNotifyScheduler 构造 NotifyScheduler。
func NewNotifyScheduler(repo *rpgrepo.RpgRepo, notify *rpgnotify.RpgNotifyService, log *zap.Logger) *NotifyScheduler {
	return &NotifyScheduler{repo: repo, notify: notify, log: log}
}

// RunDailyActivityNotify 对齐 Nest ActivityNotifyScheduler.handleDailyActivityNotify。
func (s *NotifyScheduler) RunDailyActivityNotify(ctx context.Context) error {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24*time.Hour - time.Nanosecond)

	starting, err := s.repo.ListActiveActivitiesStartingBetween(ctx, startOfDay, endOfDay)
	if err != nil {
		return err
	}
	ending, err := s.repo.ListActiveActivitiesEndingBetween(ctx, startOfDay, endOfDay)
	if err != nil {
		return err
	}

	if len(starting) > 0 {
		_ = s.notify.BroadcastActivityUpdate(ctx, rpgnotify.ActivityUpdatePayload{
			Type:       "start",
			Activities: toActivityItems(starting),
		})
		s.log.Info("activity notify start", zap.Int("count", len(starting)))
	}
	if len(ending) > 0 {
		_ = s.notify.BroadcastActivityUpdate(ctx, rpgnotify.ActivityUpdatePayload{
			Type:       "end",
			Activities: toActivityItems(ending),
		})
		s.log.Info("activity notify end", zap.Int("count", len(ending)))
	}
	return nil
}

func toActivityItems(rows []*ent.RpgActivity) []rpgnotify.ActivityItem {
	out := make([]rpgnotify.ActivityItem, 0, len(rows))
	for _, a := range rows {
		out = append(out, rpgnotify.ActivityItem{
			Code:        a.Code,
			Name:        a.Name,
			Description: a.Description,
			ExpBuffRate: a.ExpBuffRate,
		})
	}
	return out
}
