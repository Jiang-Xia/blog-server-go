// Package scheduler 定时任务框架骨架；业务 job 在 Plan 09 完善。
package scheduler

import (
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

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

// Start 启动 cron 调度。
func (s *Scheduler) Start() {
	s.cron.Start()
}

// Stop 停止 cron 调度。
func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
}
