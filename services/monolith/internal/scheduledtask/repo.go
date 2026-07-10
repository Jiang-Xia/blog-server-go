package scheduledtask

import (
	"context"
	"strings"
	"time"

	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent/scheduledtask"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent/scheduledtasklog"
)

const (
	taskStatusSuccess = "success"
	taskStatusFailed  = "failed"
)

// Repo scheduled_task / scheduled_task_log 数据访问。
type Repo struct {
	client *ent.Client
}

// NewRepo 构造 Repo。
func NewRepo(client *ent.Client) *Repo {
	return &Repo{client: client}
}

// SyncSeeds 幂等插入种子任务（已存在则跳过）。
func (r *Repo) SyncSeeds(ctx context.Context) (created int, err error) {
	for _, seed := range SeedTasks {
		exists, err := r.client.ScheduledTask.Query().
			Where(scheduledtask.NameEQ(seed.Name)).
			Exist(ctx)
		if err != nil {
			return created, err
		}
		if exists {
			continue
		}
		_, err = r.client.ScheduledTask.Create().
			SetName(seed.Name).
			SetDescription(seed.Description).
			SetCron(seed.Cron).
			SetCronHuman(seed.CronHuman).
			SetEnabled(1).
			SetLogRecording(1).
			SetSortOrder(seed.SortOrder).
			Save(ctx)
		if err != nil {
			return created, err
		}
		created++
	}
	return created, nil
}

// ListEnabled 查询已启用任务。
func (r *Repo) ListEnabled(ctx context.Context) ([]*ent.ScheduledTask, error) {
	return r.client.ScheduledTask.Query().
		Where(scheduledtask.EnabledEQ(1)).
		Order(ent.Asc(scheduledtask.FieldSortOrder)).
		All(ctx)
}

// ListAll 查询全部任务定义。
func (r *Repo) ListAll(ctx context.Context) ([]*ent.ScheduledTask, error) {
	return r.client.ScheduledTask.Query().
		Order(ent.Asc(scheduledtask.FieldSortOrder)).
		All(ctx)
}

// ListPaged 分页查询任务定义。
func (r *Repo) ListPaged(ctx context.Context, page, pageSize int, keyword string, enabled *bool) ([]*ent.ScheduledTask, int, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	q := r.client.ScheduledTask.Query()
	if keyword != "" {
		q = q.Where(
			scheduledtask.Or(
				scheduledtask.NameContains(keyword),
				scheduledtask.DescriptionContains(keyword),
			),
		)
	}
	if enabled != nil {
		v := 0
		if *enabled {
			v = 1
		}
		q = q.Where(scheduledtask.EnabledEQ(v))
	}
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	rows, err := q.Offset((page - 1) * pageSize).Limit(pageSize).
		Order(ent.Asc(scheduledtask.FieldSortOrder)).
		All(ctx)
	return rows, total, err
}

// GetByID 按 ID 查询。
func (r *Repo) GetByID(ctx context.Context, id int) (*ent.ScheduledTask, error) {
	return r.client.ScheduledTask.Query().
		Where(scheduledtask.IDEQ(id)).
		Only(ctx)
}

// GetByName 按 name 查询。
func (r *Repo) GetByName(ctx context.Context, name string) (*ent.ScheduledTask, error) {
	return r.client.ScheduledTask.Query().
		Where(scheduledtask.NameEQ(name)).
		Only(ctx)
}

// Count 任务总数。
func (r *Repo) Count(ctx context.Context) (int, error) {
	return r.client.ScheduledTask.Query().Count(ctx)
}

// Create 创建任务定义。
func (r *Repo) Create(ctx context.Context, row *ent.ScheduledTask) (*ent.ScheduledTask, error) {
	b := r.client.ScheduledTask.Create().
		SetName(row.Name).
		SetDescription(row.Description).
		SetCron(row.Cron).
		SetCronHuman(row.CronHuman).
		SetSortOrder(row.SortOrder)
	if row.Enabled != 0 {
		b.SetEnabled(1)
	}
	if row.LogRecording != 0 {
		b.SetLogRecording(1)
	}
	return b.Save(ctx)
}

// Update 更新任务定义。
func (r *Repo) Update(ctx context.Context, id int, fields map[string]interface{}) (*ent.ScheduledTask, error) {
	up := r.client.ScheduledTask.UpdateOneID(id)
	if v, ok := fields["description"].(string); ok {
		up.SetDescription(v)
	}
	if v, ok := fields["cron"].(string); ok {
		up.SetCron(v)
	}
	if v, ok := fields["cronHuman"].(string); ok {
		up.SetCronHuman(v)
	}
	if v, ok := fields["enabled"].(bool); ok {
		if v {
			up.SetEnabled(1)
		} else {
			up.SetEnabled(0)
		}
	}
	if v, ok := fields["logRecording"].(bool); ok {
		if v {
			up.SetLogRecording(1)
		} else {
			up.SetLogRecording(0)
		}
	}
	if v, ok := fields["sortOrder"].(int); ok {
		up.SetSortOrder(v)
	}
	return up.Save(ctx)
}

// Delete 硬删除任务定义（Nest 表无软删除列）。
func (r *Repo) Delete(ctx context.Context, id int) error {
	return r.client.ScheduledTask.DeleteOneID(id).Exec(ctx)
}

// SetEnabled 设置启用状态。
func (r *Repo) SetEnabled(ctx context.Context, name string, enabled bool) (*ent.ScheduledTask, error) {
	row, err := r.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}
	v := 0
	if enabled {
		v = 1
	}
	return r.client.ScheduledTask.UpdateOneID(row.ID).SetEnabled(v).Save(ctx)
}

// ToggleLogRecording 切换日志记录开关。
func (r *Repo) ToggleLogRecording(ctx context.Context, name string) (*ent.ScheduledTask, error) {
	row, err := r.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}
	v := 1
	if row.LogRecording != 0 {
		v = 0
	}
	return r.client.ScheduledTask.UpdateOneID(row.ID).SetLogRecording(v).Save(ctx)
}

// LogFilter 执行日志查询条件。
type LogFilter struct {
	Page     int
	PageSize int
	TaskName string
	Status   string
}

// ListLogs 分页查询执行日志。
func (r *Repo) ListLogs(ctx context.Context, f LogFilter) ([]*ent.ScheduledTaskLog, int, error) {
	page, pageSize := f.Page, f.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	q := r.client.ScheduledTaskLog.Query()
	if f.TaskName != "" {
		q = q.Where(scheduledtasklog.TaskNameEQ(f.TaskName))
	}
	if f.Status != "" {
		q = q.Where(scheduledtasklog.StatusEQ(f.Status))
	}
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	rows, err := q.Order(ent.Desc(scheduledtasklog.FieldCreateTime)).
		Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	return rows, total, err
}

// SaveLog 写入执行日志（受 logRecording 控制）。
func (r *Repo) SaveLog(ctx context.Context, taskName, status string, start time.Time, result, errMsg string) error {
	task, err := r.GetByName(ctx, taskName)
	if err == nil && task.LogRecording == 0 {
		return nil
	}
	now := time.Now()
	b := r.client.ScheduledTaskLog.Create().
		SetTaskName(taskName).
		SetStatus(status).
		SetStartTime(start).
		SetEndTime(now)
	if result != "" {
		b.SetResult(result)
	}
	if errMsg != "" {
		b.SetErrorMessage(errMsg)
	}
	_, err = b.Save(ctx)
	return err
}

// DeleteOldLogs 删除过期任务日志。
func (r *Repo) DeleteOldLogs(ctx context.Context, before time.Time) (int, error) {
	n, err := r.client.ScheduledTaskLog.Delete().
		Where(
			scheduledtasklog.CreateTimeLT(before),
		).
		Exec(ctx)
	return n, err
}

// DeleteOldOperationLogs 已由 crossdb 包在 expired_data_cleanup job 中处理。

// TaskDTO 对外任务定义（含 running）。
func TaskDTO(row *ent.ScheduledTask, running bool) map[string]interface{} {
	return map[string]interface{}{
		"id":            row.ID,
		"name":          row.Name,
		"description":   row.Description,
		"cron":          row.Cron,
		"cronHuman":     row.CronHuman,
		"enabled":       row.Enabled != 0,
		"logRecording":  row.LogRecording != 0,
		"sortOrder":     row.SortOrder,
		"createTime":    row.CreateTime,
		"updateTime":    row.UpdateTime,
		"running":       running,
	}
}

// ParseEnabledQuery 解析 enabled 查询参数。
func ParseEnabledQuery(raw string) *bool {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	v := strings.EqualFold(raw, "true") || raw == "1"
	return &v
}
