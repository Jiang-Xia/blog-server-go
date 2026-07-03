// scheduled_task_handler 定时任务与运维 HTTP 端点（Plan 12）。
package handler

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/Jiang-Xia/blog-server-go/pkg/ctxutil"
	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/pkg/response"
	"github.com/Jiang-Xia/blog-server-go/services/blog/ent"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/auth"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/scheduledtask"
	"github.com/cloudwego/hertz/pkg/app"
)

// ScheduledTaskHandler 定时任务 HTTP 端点。
type ScheduledTaskHandler struct {
	svc *scheduledtask.Service
	jwt *auth.JWTService
}

// NewScheduledTaskHandler 构造 ScheduledTaskHandler。
func NewScheduledTaskHandler(svc *scheduledtask.Service, jwt *auth.JWTService) *ScheduledTaskHandler {
	return &ScheduledTaskHandler{svc: svc, jwt: jwt}
}

func (h *ScheduledTaskHandler) ListTasks(ctx context.Context, c *app.RequestContext) {
	data, err := h.svc.ListTasks(ctx)
	handleAdminResult(ctx, c, data, err)
}

func (h *ScheduledTaskHandler) ListTasksPaged(ctx context.Context, c *app.RequestContext) {
	page, _ := strconv.Atoi(string(c.Query("page")))
	pageSize, _ := strconv.Atoi(string(c.Query("pageSize")))
	data, err := h.svc.ListTasksPaged(ctx, page, pageSize, string(c.Query("keyword")), string(c.Query("enabled")))
	handleAdminResult(ctx, c, data, err)
}

func (h *ScheduledTaskHandler) GetTask(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.Param("id"))
	data, err := h.svc.GetTask(ctx, id)
	handleAdminResult(ctx, c, data, err)
}

func (h *ScheduledTaskHandler) CreateTask(ctx context.Context, c *app.RequestContext) {
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	row := &ent.ScheduledTask{
		Name:        strField(body, "name"),
		Description: strField(body, "description"),
		Cron:        strField(body, "cron"),
		CronHuman:   strField(body, "cronHuman"),
		SortOrder:   intField(body, "sortOrder"),
	}
	if v, ok := body["enabled"].(bool); ok && v {
		row.Enabled = 1
	}
	if v, ok := body["logRecording"].(bool); ok && v {
		row.LogRecording = 1
	}
	data, err := h.svc.CreateTask(ctx, row)
	handleAdminResult(ctx, c, data, err)
}

func (h *ScheduledTaskHandler) UpdateTask(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.Param("id"))
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	fields := map[string]interface{}{}
	for _, k := range []string{"description", "cron", "cronHuman"} {
		if body[k] != nil {
			fields[k] = strField(body, k)
		}
	}
	if body["enabled"] != nil {
		fields["enabled"] = body["enabled"] == true
	}
	if body["logRecording"] != nil {
		fields["logRecording"] = body["logRecording"] == true
	}
	if body["sortOrder"] != nil {
		fields["sortOrder"] = intField(body, "sortOrder")
	}
	data, err := h.svc.UpdateTask(ctx, id, fields)
	handleAdminResult(ctx, c, data, err)
}

func (h *ScheduledTaskHandler) DeleteTask(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.Param("id"))
	handleAdminResult(ctx, c, nil, h.svc.DeleteTask(ctx, id))
}

func (h *ScheduledTaskHandler) GetStatus(ctx context.Context, c *app.RequestContext) {
	handleAdminResult(ctx, c, h.svc.GetStatus(ctx, c.Param("taskName")), nil)
}

func (h *ScheduledTaskHandler) Trigger(ctx context.Context, c *app.RequestContext) {
	result, err := h.svc.TriggerTask(ctx, c.Param("taskName"))
	if err != nil {
		handleAdminResult(ctx, c, nil, err)
		return
	}
	response.Success(ctx, c, result)
}

func (h *ScheduledTaskHandler) Stop(ctx context.Context, c *app.RequestContext) {
	data, err := h.svc.StopTask(ctx, c.Param("taskName"))
	handleAdminResult(ctx, c, data, err)
}

func (h *ScheduledTaskHandler) Start(ctx context.Context, c *app.RequestContext) {
	data, err := h.svc.StartTask(ctx, c.Param("taskName"))
	handleAdminResult(ctx, c, data, err)
}

func (h *ScheduledTaskHandler) ToggleLogRecording(ctx context.Context, c *app.RequestContext) {
	data, err := h.svc.ToggleLogRecording(ctx, c.Param("taskName"))
	handleAdminResult(ctx, c, data, err)
}

func (h *ScheduledTaskHandler) ClearPermissionCache(ctx context.Context, c *app.RequestContext) {
	uid := ctxutil.UserID(ctx)
	if uid == 0 {
		uid = articleUID(ctx, c, h.jwt)
	}
	data, err := h.svc.ClearPermissionCache(ctx, uid)
	handleAdminResult(ctx, c, data, err)
}

func (h *ScheduledTaskHandler) RefreshTongjiToken(ctx context.Context, c *app.RequestContext) {
	data, err := h.svc.RefreshTongjiToken(ctx)
	handleAdminResult(ctx, c, data, err)
}

func (h *ScheduledTaskHandler) ListLogs(ctx context.Context, c *app.RequestContext) {
	page, _ := strconv.Atoi(string(c.Query("page")))
	pageSize, _ := strconv.Atoi(string(c.Query("pageSize")))
	f := scheduledtask.LogFilter{
		Page: page, PageSize: pageSize,
		TaskName: string(c.Query("taskName")),
		Status:   string(c.Query("status")),
	}
	data, err := h.svc.ListLogs(ctx, f)
	handleAdminResult(ctx, c, data, err)
}

func (h *ScheduledTaskHandler) ListBackups(ctx context.Context, c *app.RequestContext) {
	data, err := h.svc.ListBackups(ctx)
	handleAdminResult(ctx, c, data, err)
}

func (h *ScheduledTaskHandler) DownloadLatestBackup(ctx context.Context, c *app.RequestContext) {
	path, fileName, err := h.svc.ResolveLatestBackup(ctx)
	if err != nil {
		handleAdminResult(ctx, c, nil, err)
		return
	}
	serveBackupFile(ctx, c, path, fileName)
}

func (h *ScheduledTaskHandler) DownloadBackup(ctx context.Context, c *app.RequestContext) {
	fileName := c.Param("fileName")
	path, err := h.svc.ResolveBackupDownload(ctx, fileName)
	if err != nil {
		handleAdminResult(ctx, c, nil, err)
		return
	}
	serveBackupFile(ctx, c, path, fileName)
}

func serveBackupFile(ctx context.Context, c *app.RequestContext, path, fileName string) {
	data, err := os.ReadFile(path)
	if err != nil {
		response.FromError(ctx, c, errcode.WithMessage(errcode.InternalError, "读取备份文件失败"))
		return
	}
	c.Header("Content-Type", "application/sql")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileName))
	c.SetStatusCode(200)
	c.Write(data)
}
