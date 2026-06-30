// Package scheduler 定时任务框架；Plan 09 接入 RPG 活动通知等 job。
package scheduler

import (
	"context"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// ActivityNotifyRunner 每日活动开始/结束推送任务。
type ActivityNotifyRunner interface {
	RunDailyActivityNotify(ctx context.Context) error
}

// Scheduler robfig/cron 封装，应用启动时注册占位 job。
type Scheduler struct {
	cron *cron.Cron
	log  *zap.Logger
}

// New 构造调度器（秒级 cron 表达式，与 Nest 6 段一致）。
func New(log *zap.Logger) *Scheduler {
	return &Scheduler{
		cron: cron.New(cron.WithSeconds()),
		log:  log,
	}
}

// RegisterPlaceholder 注册占位 job（每小时整点打 debug 日志）。
func (s *Scheduler) RegisterPlaceholder() error {
	_, err := s.cron.AddFunc("0 0 * * * *", func() {
		s.log.Debug("placeholder cron job tick")
	})
	return err
}

// RegisterActivityNotify 注册每日 8:00 活动通知 job（对齐 Nest ActivityNotifyScheduler）。
func (s *Scheduler) RegisterActivityNotify(runner ActivityNotifyRunner) error {
	if runner == nil {
		return nil
	}
	_, err := s.cron.AddFunc("0 0 8 * * *", func() {
		ctx := context.Background()
		if err := runner.RunDailyActivityNotify(ctx); err != nil {
			s.log.Warn("activity notify job failed", zap.Error(err))
		}
	})
	return err
}

// Start 启动 cron 调度。
func (s *Scheduler) Start() {
	s.cron.Start()
}

// Stop 停止 cron 调度。
func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
}
