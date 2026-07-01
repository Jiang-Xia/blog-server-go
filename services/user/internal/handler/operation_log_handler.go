package handler

import (
	"context"
	"strconv"

	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/pkg/response"
	"github.com/Jiang-Xia/blog-server-go/services/user/internal/operationlog"
	"github.com/cloudwego/hertz/pkg/app"
)

// OperationLogHandler 操作日志 HTTP 端点。
type OperationLogHandler struct {
	svc *operationlog.Service
}

// NewOperationLogHandler 构造 OperationLogHandler。
func NewOperationLogHandler(svc *operationlog.Service) *OperationLogHandler {
	return &OperationLogHandler{svc: svc}
}

func (h *OperationLogHandler) List(ctx context.Context, c *app.RequestContext) {
	page, _ := strconv.Atoi(string(c.Query("page")))
	pageSize, _ := strconv.Atoi(string(c.Query("pageSize")))
	data, err := h.svc.List(ctx, operationlog.ListQuery{
		Page:     page,
		PageSize: pageSize,
		Module:   string(c.Query("module")),
		Action:   string(c.Query("action")),
		Username: string(c.Query("username")),
		Keyword:  string(c.Query("keyword")),
	})
	handleAdminResult(ctx, c, data, err)
}

func (h *OperationLogHandler) Clean(ctx context.Context, c *app.RequestContext) {
	days := 0
	if s := string(c.Query("days")); s != "" {
		v, err := strconv.Atoi(s)
		if err != nil {
			response.Error(ctx, c, errcode.InvalidParam)
			return
		}
		days = v
	}
	n, err := h.svc.Clean(ctx, days)
	handleAdminResult(ctx, c, map[string]int{"deleted": n}, err)
}
