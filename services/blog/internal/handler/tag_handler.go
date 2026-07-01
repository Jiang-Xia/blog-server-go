// tag_handler 标签 HTTP 端点。
package handler

import (
	"context"
	"strings"

	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/pkg/response"
	blogsvc "github.com/Jiang-Xia/blog-server-go/services/blog/internal/blog/service"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/auth"
	"github.com/cloudwego/hertz/pkg/app"
)

// TagHandler 标签 HTTP 端点。
type TagHandler struct {
	svc *blogsvc.TagService
	jwt *auth.JWTService
}

// NewTagHandler 构造 TagHandler。
func NewTagHandler(svc *blogsvc.TagService, jwt *auth.JWTService) *TagHandler {
	return &TagHandler{svc: svc, jwt: jwt}
}

func (h *TagHandler) Create(ctx context.Context, c *app.RequestContext) {
	uid := articleUID(ctx, c, h.jwt)
	if uid == 0 {
		response.Error(ctx, c, errcode.WithMessage(errcode.Unauthorized, "身份验证失败"))
		return
	}
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	data, err := h.svc.Create(ctx, uid, strField(body, "label"), strField(body, "value"))
	handleAdminResult(ctx, c, data, err)
}

func (h *TagHandler) List(ctx context.Context, c *app.RequestContext) {
	title := string(c.Query("title"))
	countPublished := strings.EqualFold(string(c.Query("isDelete")), "true")
	data, err := h.svc.List(ctx, title, countPublished)
	handleAdminResult(ctx, c, data, err)
}

func (h *TagHandler) Get(ctx context.Context, c *app.RequestContext) {
	data, err := h.svc.Get(ctx, c.Param("id"))
	handleAdminResult(ctx, c, data, err)
}

func (h *TagHandler) GetArticles(ctx context.Context, c *app.RequestContext) {
	status := string(c.Query("status"))
	data, err := h.svc.GetWithArticles(ctx, c.Param("id"), status)
	handleAdminResult(ctx, c, data, err)
}

func (h *TagHandler) Update(ctx context.Context, c *app.RequestContext) {
	uid := articleUID(ctx, c, h.jwt)
	if uid == 0 {
		response.Error(ctx, c, errcode.WithMessage(errcode.Unauthorized, "身份验证失败"))
		return
	}
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	var label, value, color *string
	if v, ok := body["label"].(string); ok {
		label = &v
	}
	if v, ok := body["value"].(string); ok {
		value = &v
	}
	if v, ok := body["color"].(string); ok {
		color = &v
	}
	data, err := h.svc.Update(ctx, c.Param("id"), label, value, color)
	handleAdminResult(ctx, c, data, err)
}

func (h *TagHandler) Delete(ctx context.Context, c *app.RequestContext) {
	uid := articleUID(ctx, c, h.jwt)
	if uid == 0 {
		response.Error(ctx, c, errcode.WithMessage(errcode.Unauthorized, "身份验证失败"))
		return
	}
	err := h.svc.Delete(ctx, c.Param("id"))
	handleAdminResult(ctx, c, true, err)
}
