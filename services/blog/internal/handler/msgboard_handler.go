// msgboard_handler 留言板 HTTP 端点。
package handler

import (
	"context"

	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/pkg/response"
	blogsvc "github.com/Jiang-Xia/blog-server-go/services/blog/internal/blog/service"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/auth"
	"github.com/cloudwego/hertz/pkg/app"
)

// MsgboardHandler 留言板 HTTP 端点。
type MsgboardHandler struct {
	svc *blogsvc.MsgboardService
	jwt *auth.JWTService
}

// NewMsgboardHandler 构造 MsgboardHandler。
func NewMsgboardHandler(svc *blogsvc.MsgboardService, jwt *auth.JWTService) *MsgboardHandler {
	return &MsgboardHandler{svc: svc, jwt: jwt}
}

func (h *MsgboardHandler) Create(ctx context.Context, c *app.RequestContext) {
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	ip := string(c.ClientIP())
	ua := string(c.UserAgent())
	data, err := h.svc.Create(ctx, body, ip, ua)
	handleAdminResult(ctx, c, data, err)
}

func (h *MsgboardHandler) List(ctx context.Context, c *app.RequestContext) {
	q := map[string]interface{}{}
	_ = c.Bind(&q)
	if q["page"] == nil {
		q["page"] = string(c.Query("page"))
	}
	if q["pageSize"] == nil {
		q["pageSize"] = string(c.Query("pageSize"))
	}
	if q["status"] == nil {
		q["status"] = string(c.Query("status"))
	}
	data, err := h.svc.List(ctx, q)
	handleAdminResult(ctx, c, data, err)
}

func (h *MsgboardHandler) Delete(ctx context.Context, c *app.RequestContext) {
	var ids []int
	if err := c.Bind(&ids); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	handleAdminResult(ctx, c, nil, h.svc.Delete(ctx, ids))
}
