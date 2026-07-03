// Package scheduler robfig/cron 封装，按 DB scheduled_task 动态注册 job。
package scheduler

import (
	"context"
	"sync"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// SchedulerControl cron 注册/注销接口。
type SchedulerControl interface {
	RegisterTask(name, cronExpr string) error
	UnregisterTask(name string)
	IsRunning(name string) bool
}

// TaskTrigger 任务触发回调。
type TaskTrigger func(ctx context.Context, taskName string) (interface{}, error)

// Scheduler cron 调度器，实现 scheduledtask.SchedulerControl。
type Scheduler struct {
	cron    *cron.Cron
	log     *zap.Logger
	entries map[string]cron.EntryID
	mu      sync.Mutex
	trigger TaskTrigger
}

// New 构造调度器（秒级 cron 表达式，与 Nest 6 段一致）。
func New(log *zap.Logger) *Scheduler {
	return &Scheduler{
		cron:    cron.New(cron.WithSeconds()),
		log:     log,
		entries: make(map[string]cron.EntryID),
	}
}

// SetTrigger 设置任务触发回调（由 Service 注入）。
func (s *Scheduler) SetTrigger(fn TaskTrigger) {
	s.trigger = fn
}

// RegisterTask 注册或替换单个 cron job。
func (s *Scheduler) RegisterTask(name, cronExpr string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.removeLocked(name)
	id, err := s.cron.AddFunc(cronExpr, func() {
		if s.trigger == nil {
			return
		}
		ctx := context.Background()
		if _, err := s.trigger(ctx, name); err != nil && s.log != nil {
			s.log.Warn("cron task failed", zap.String("task", name), zap.Error(err))
		}
	})
	if err != nil {
		return err
	}
	s.entries[name] = id
	return nil
}

// UnregisterTask 注销 cron job。
func (s *Scheduler) UnregisterTask(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.removeLocked(name)
}

// IsRunning 判断 job 是否已注册（Ent enabled 且 cron 中存在）。
func (s *Scheduler) IsRunning(name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.entries[name]
	return ok
}

func (s *Scheduler) removeLocked(name string) {
	if id, ok := s.entries[name]; ok {
		s.cron.Remove(id)
		delete(s.entries, name)
	}
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
