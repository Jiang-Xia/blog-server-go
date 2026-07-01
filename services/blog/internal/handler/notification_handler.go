// notification_handler 站内通知 HTTP 端点。
package handler

import (
	"context"
	"strconv"
	"strings"

	"github.com/Jiang-Xia/blog-server-go/pkg/ctxutil"
	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/pkg/response"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/blog/notification"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/auth"
	"github.com/cloudwego/hertz/pkg/app"
)

// NotificationHandler 站内通知 HTTP 端点。
type NotificationHandler struct {
	svc *notification.Service
	jwt *auth.JWTService
}

// NewNotificationHandler 构造 NotificationHandler。
func NewNotificationHandler(svc *notification.Service, jwt *auth.JWTService) *NotificationHandler {
	return &NotificationHandler{svc: svc, jwt: jwt}
}

func (h *NotificationHandler) List(ctx context.Context, c *app.RequestContext) {
	uid := h.requireUID(ctx, c)
	if uid == 0 {
		response.Error(ctx, c, errcode.Unauthorized)
		return
	}
	page, _ := strconv.Atoi(string(c.Query("page")))
	pageSize, _ := strconv.Atoi(string(c.Query("pageSize")))
	data, err := h.svc.ListByUID(ctx, uid, page, pageSize)
	handleAdminResult(ctx, c, data, err)
}

func (h *NotificationHandler) UnreadCount(ctx context.Context, c *app.RequestContext) {
	uid := h.requireUID(ctx, c)
	if uid == 0 {
		response.Error(ctx, c, errcode.Unauthorized)
		return
	}
	count, err := h.svc.CountUnread(ctx, uid)
	handleAdminResult(ctx, c, count, err)
}

func (h *NotificationHandler) MarkRead(ctx context.Context, c *app.RequestContext) {
	uid := h.requireUID(ctx, c)
	if uid == 0 {
		response.Error(ctx, c, errcode.Unauthorized)
		return
	}
	var idPtr *int
	if s := string(c.Query("id")); s != "" {
		v, err := strconv.Atoi(s)
		if err == nil {
			idPtr = &v
		}
	}
	err := h.svc.MarkRead(ctx, uid, idPtr)
	handleAdminResult(ctx, c, nil, err)
}

// Since GET /notification/since?seq= 骨架（Plan 08 完善 WS 补漏）。
func (h *NotificationHandler) Since(ctx context.Context, c *app.RequestContext) {
	uid := h.requireUID(ctx, c)
	if uid == 0 {
		response.Error(ctx, c, errcode.Unauthorized)
		return
	}
	seq, _ := strconv.Atoi(string(c.Query("seq")))
	list, err := h.svc.Since(ctx, uid, seq)
	handleAdminResult(ctx, c, list, err)
}

func (h *NotificationHandler) requireUID(ctx context.Context, c *app.RequestContext) int {
	if uid := ctxutil.UserID(ctx); uid != 0 {
		return uid
	}
	if h.jwt == nil {
		return 0
	}
	authz := strings.TrimSpace(string(c.GetHeader("Authorization")))
	if authz == "" {
		return 0
	}
	token := strings.TrimPrefix(authz, "Bearer ")
	claims, err := h.jwt.Verify(strings.TrimSpace(token))
	if err != nil || claims == nil {
		return 0
	}
	return claims.ID
}
