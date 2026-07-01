// link_handler 友链 HTTP 端点。
package handler

import (
	"context"
	"strconv"

	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/pkg/response"
	blogsvc "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/blog/service"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/auth"
	"github.com/cloudwego/hertz/pkg/app"
)

// LinkHandler 友链 HTTP 端点。
type LinkHandler struct {
	svc *blogsvc.LinkService
	jwt *auth.JWTService
}

// NewLinkHandler 构造 LinkHandler。
func NewLinkHandler(svc *blogsvc.LinkService, jwt *auth.JWTService) *LinkHandler {
	return &LinkHandler{svc: svc, jwt: jwt}
}

func (h *LinkHandler) Create(ctx context.Context, c *app.RequestContext) {
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	data, err := h.svc.Create(ctx, strField(body, "icon"), strField(body, "url"), strField(body, "title"), strField(body, "desp"))
	handleAdminResult(ctx, c, data, err)
}

func (h *LinkHandler) Get(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.Param("id"))
	data, err := h.svc.Get(ctx, id)
	handleAdminResult(ctx, c, data, err)
}

func (h *LinkHandler) List(ctx context.Context, c *app.RequestContext) {
	client := string(c.Query("client")) != ""
	data, err := h.svc.List(ctx, client, string(c.Query("title")), string(c.Query("url")), string(c.Query("agreed")))
	handleAdminResult(ctx, c, data, err)
}

func (h *LinkHandler) Update(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(c.Param("id"))
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	data, err := h.svc.Update(ctx, id, body)
	handleAdminResult(ctx, c, data, err)
}

func (h *LinkHandler) Delete(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(string(c.Query("id")))
	handleAdminResult(ctx, c, nil, h.svc.Delete(ctx, id))
}
