package scheduledtask

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/ctxutil"
	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/pkg/pagination"
	"github.com/Jiang-Xia/blog-server-go/pkg/redisutil"
	"github.com/Jiang-Xia/blog-server-go/services/blog/ent"
	blogsvc "github.com/Jiang-Xia/blog-server-go/services/blog/internal/blog/service"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/blog/scheduler"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/crossdb"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/scheduledtask/jobs"
	"go.uber.org/zap"
)

const (
	lockKeyPrefix  = "scheduled_task:lock:"
	lockTTLSeconds = 300
)

// Service 定时任务 CRUD、调度触发与运维能力。
type Service struct {
	repo   *Repo
	cross  *crossdb.CrossDB
	cfg    *config.Config
	redis  *redisutil.Store
	jobs   *jobs.Runner
	log    *zap.Logger
	sched  scheduler.SchedulerControl
	tongji blogsvc.TongjiRefresher
	mu     sync.Map // taskName -> *sync.Mutex 本地互斥
	instID string
}

// NewService 构造 Service。
func NewService(
	repo *Repo,
	cross *crossdb.CrossDB,
	cfg *config.Config,
	redis *redisutil.Store,
	runner *jobs.Runner,
	log *zap.Logger,
	tongji blogsvc.TongjiRefresher,
) *Service {
	return &Service{
		repo: repo, cross: cross, cfg: cfg, redis: redis, jobs: runner, log: log, tongji: tongji,
		instID: fmt.Sprintf("%d", os.Getpid()),
	}
}

// SetScheduler 注入 cron 控制器（避免 wire 循环依赖）。
func (s *Service) SetScheduler(ctrl scheduler.SchedulerControl) {
	s.sched = ctrl
}

// Bootstrap 种子同步 + 注册已启用 cron。
func (s *Service) Bootstrap(ctx context.Context) error {
	n, err := s.repo.SyncSeeds(ctx)
	if err != nil {
		return err
	}
	if n > 0 && s.log != nil {
		s.log.Info("scheduled task seeds created", zap.Int("count", n))
	}
	return s.ReloadCron(ctx)
}

// ReloadCron 从 DB 重新加载全部启用任务到 cron。
func (s *Service) ReloadCron(ctx context.Context) error {
	if s.sched == nil {
		return nil
	}
	tasks, err := s.repo.ListEnabled(ctx)
	if err != nil {
		return err
	}
	for _, t := range tasks {
		if err := s.sched.RegisterTask(t.Name, t.Cron); err != nil && s.log != nil {
			s.log.Error("register cron failed", zap.String("task", t.Name), zap.Error(err))
		}
	}
	if s.log != nil {
		s.log.Info("scheduled tasks registered", zap.Int("count", len(tasks)))
	}
	return nil
}

// ListTasks 全部任务（含 running）。
func (s *Service) ListTasks(ctx context.Context) ([]map[string]interface{}, error) {
	rows, err := s.repo.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		running := row.Enabled != 0 && s.isRunning(row.Name)
		out = append(out, TaskDTO(row, running))
	}
	return out, nil
}

// ListTasksPaged 分页任务定义。
func (s *Service) ListTasksPaged(ctx context.Context, page, pageSize int, keyword, enabledRaw string) (map[string]interface{}, error) {
	rows, total, err := s.repo.ListPaged(ctx, page, pageSize, keyword, ParseEnabledQuery(enabledRaw))
	if err != nil {
		return nil, err
	}
	list := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		list = append(list, TaskDTO(row, row.Enabled != 0 && s.isRunning(row.Name)))
	}
	return map[string]interface{}{
		"list":       list,
		"pagination": pagination.CalcNestPagination(total, pageSize, page),
	}, nil
}

// GetTask 单个任务。
func (s *Service) GetTask(ctx context.Context, id int) (map[string]interface{}, error) {
	row, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errcode.WithMessage(errcode.InvalidParam, "任务不存在: id=%d", id)
		}
		return nil, err
	}
	return TaskDTO(row, row.Enabled != 0 && s.isRunning(row.Name)), nil
}

// CreateTask 创建任务。
func (s *Service) CreateTask(ctx context.Context, row *ent.ScheduledTask) (*ent.ScheduledTask, error) {
	if _, err := s.repo.GetByName(ctx, row.Name); err == nil {
		return nil, errcode.WithMessage(errcode.InvalidParam, "任务标识已存在: %s", row.Name)
	} else if !ent.IsNotFound(err) {
		return nil, err
	}
	count, err := s.repo.Count(ctx)
	if err != nil {
		return nil, err
	}
	if count >= maxTasks {
		return nil, errcode.WithMessage(errcode.InvalidParam, "任务数已达上限（%d）", maxTasks)
	}
	saved, err := s.repo.Create(ctx, row)
	if err != nil {
		return nil, err
	}
	if saved.Enabled != 0 && s.sched != nil {
		_ = s.sched.RegisterTask(saved.Name, saved.Cron)
	}
	return saved, nil
}

// UpdateTask 更新任务。
func (s *Service) UpdateTask(ctx context.Context, id int, fields map[string]interface{}) (*ent.ScheduledTask, error) {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		if ent.IsNotFound(err) {
			return nil, errcode.WithMessage(errcode.InvalidParam, "任务不存在: id=%d", id)
		}
		return nil, err
	}
	saved, err := s.repo.Update(ctx, id, fields)
	if err != nil {
		return nil, err
	}
	if s.sched != nil {
		s.sched.UnregisterTask(saved.Name)
		if saved.Enabled != 0 {
			_ = s.sched.RegisterTask(saved.Name, saved.Cron)
		}
	}
	return saved, nil
}

// DeleteTask 删除任务。
func (s *Service) DeleteTask(ctx context.Context, id int) error {
	row, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return errcode.WithMessage(errcode.InvalidParam, "任务不存在: id=%d", id)
		}
		return err
	}
	if s.sched != nil {
		s.sched.UnregisterTask(row.Name)
	}
	return s.repo.Delete(ctx, id)
}

// GetStatus 任务运行状态。
func (s *Service) GetStatus(_ context.Context, taskName string) map[string]interface{} {
	return map[string]interface{}{
		"taskName": taskName,
		"running":  s.isRunning(taskName),
		"lastDate": nil,
	}
}

// TriggerTask 手动触发（带 Redis 锁 + 本地互斥）。
func (s *Service) TriggerTask(ctx context.Context, taskName string) (interface{}, error) {
	acquired, err := s.tryAcquireLock(ctx, taskName)
	if err != nil {
		return nil, err
	}
	if !acquired {
		return map[string]interface{}{"skipped": true, "reason": "另一个实例正在执行该任务"}, nil
	}
	defer s.releaseLock(ctx, taskName)

	mu := s.taskMutex(taskName)
	mu.Lock()
	defer mu.Unlock()

	return s.executeTask(ctx, taskName)
}

// StopTask 停止 cron。
func (s *Service) StopTask(ctx context.Context, taskName string) (map[string]interface{}, error) {
	if _, err := s.repo.GetByName(ctx, taskName); err != nil {
		if ent.IsNotFound(err) {
			return nil, errcode.WithMessage(errcode.InvalidParam, "未知的任务名称: %s", taskName)
		}
		return nil, err
	}
	if _, err := s.repo.SetEnabled(ctx, taskName, false); err != nil {
		return nil, err
	}
	if s.sched != nil {
		s.sched.UnregisterTask(taskName)
	}
	return map[string]interface{}{"taskName": taskName, "running": false}, nil
}

// StartTask 启动 cron。
func (s *Service) StartTask(ctx context.Context, taskName string) (map[string]interface{}, error) {
	row, err := s.repo.SetEnabled(ctx, taskName, true)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errcode.WithMessage(errcode.InvalidParam, "未知的任务名称: %s", taskName)
		}
		return nil, err
	}
	if s.sched != nil {
		_ = s.sched.RegisterTask(row.Name, row.Cron)
	}
	return map[string]interface{}{"taskName": taskName, "running": true}, nil
}

// ToggleLogRecording 切换日志开关。
func (s *Service) ToggleLogRecording(ctx context.Context, taskName string) (map[string]interface{}, error) {
	row, err := s.repo.ToggleLogRecording(ctx, taskName)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errcode.WithMessage(errcode.InvalidParam, "未知的任务名称: %s", taskName)
		}
		return nil, err
	}
	return map[string]interface{}{"taskName": taskName, "logRecording": row.LogRecording != 0}, nil
}

// ListLogs 分页执行日志。
func (s *Service) ListLogs(ctx context.Context, f LogFilter) (map[string]interface{}, error) {
	rows, total, err := s.repo.ListLogs(ctx, f)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"list":       rows,
		"pagination": pagination.CalcNestPagination(total, f.PageSize, f.Page),
	}, nil
}

// ClearPermissionCache 超管清除 RBAC Redis 缓存。
func (s *Service) ClearPermissionCache(ctx context.Context, uid int) (map[string]interface{}, error) {
	if !ctxutil.IsSuperAdmin(ctx) {
		return nil, errcode.WithMessage(errcode.Forbidden, "仅超级管理员可操作")
	}
	_ = uid
	deleted := 0
	if err := s.redis.Del(ctx, "public_api_paths", "api_permission_mappings"); err != nil {
		return nil, err
	}
	deleted += 2
	n, err := s.redis.DelByPattern(ctx, "role_permissions:*")
	if err != nil {
		return nil, err
	}
	deleted += n
	return map[string]interface{}{"deleted": deleted}, nil
}

// RefreshTongjiToken 超管刷新百度 token，委托 ResourcesService。
func (s *Service) RefreshTongjiToken(ctx context.Context) (map[string]interface{}, error) {
	if !ctxutil.IsSuperAdmin(ctx) {
		return nil, errcode.WithMessage(errcode.Forbidden, "仅超级管理员可操作")
	}
	if s.tongji == nil {
		return nil, errcode.WithMessage(errcode.InternalError, "百度统计服务未就绪")
	}
	return s.tongji.ForceRefreshTongjiAccessToken(ctx)
}

// ListBackups 超管列出备份文件。
func (s *Service) ListBackups(ctx context.Context) ([]BackupFileInfo, error) {
	if !ctxutil.IsSuperAdmin(ctx) {
		return nil, errcode.WithMessage(errcode.Forbidden, "仅超级管理员可操作")
	}
	return ListBackupFiles(s.cfg)
}

// ResolveLatestBackup 最新备份路径。
func (s *Service) ResolveLatestBackup(ctx context.Context) (string, string, error) {
	if !ctxutil.IsSuperAdmin(ctx) {
		return "", "", errcode.WithMessage(errcode.Forbidden, "仅超级管理员可操作")
	}
	files, err := ListBackupFiles(s.cfg)
	if err != nil {
		return "", "", err
	}
	if len(files) == 0 {
		return "", "", errcode.WithMessage(errcode.NotFound, "暂无备份文件")
	}
	path, err := ResolveBackupPath(s.cfg, files[0].FileName)
	return path, files[0].FileName, err
}

// ResolveBackupDownload 指定备份路径。
func (s *Service) ResolveBackupDownload(ctx context.Context, fileName string) (string, error) {
	if !ctxutil.IsSuperAdmin(ctx) {
		return "", errcode.WithMessage(errcode.Forbidden, "仅超级管理员可操作")
	}
	return ResolveBackupPath(s.cfg, fileName)
}

func (s *Service) executeTask(ctx context.Context, taskName string) (interface{}, error) {
	start := time.Now()
	defer func() {
		if r := recover(); r != nil {
			msg := fmt.Sprint(r)
			_ = s.repo.SaveLog(ctx, taskName, taskStatusFailed, start, "", msg)
			if s.log != nil {
				s.log.Error("scheduled task panic", zap.String("task", taskName), zap.Any("panic", r))
			}
		}
	}()

	result, err := s.jobs.Execute(ctx, taskName)
	if err != nil {
		_ = s.repo.SaveLog(ctx, taskName, taskStatusFailed, start, "", err.Error())
		return nil, err
	}
	raw, _ := json.Marshal(result)
	_ = s.repo.SaveLog(ctx, taskName, taskStatusSuccess, start, string(raw), "")
	return result, nil
}

func (s *Service) isRunning(name string) bool {
	if s.sched == nil {
		return false
	}
	return s.sched.IsRunning(name)
}

func (s *Service) tryAcquireLock(ctx context.Context, taskName string) (bool, error) {
	if s.redis == nil {
		return true, nil
	}
	return s.redis.TryAcquireLock(ctx, lockKeyPrefix+taskName, s.instID, lockTTLSeconds)
}

func (s *Service) releaseLock(ctx context.Context, taskName string) {
	if s.redis == nil {
		return
	}
	_ = s.redis.ReleaseLock(ctx, lockKeyPrefix+taskName, s.instID)
}

func (s *Service) taskMutex(taskName string) *sync.Mutex {
	v, _ := s.mu.LoadOrStore(taskName, &sync.Mutex{})
	return v.(*sync.Mutex)
}
