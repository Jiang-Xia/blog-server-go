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

// CategoryHandler 分类 HTTP 端点。
type CategoryHandler struct {
	svc *blogsvc.CategoryService
	jwt *auth.JWTService
}

// NewCategoryHandler 构造 CategoryHandler。
func NewCategoryHandler(svc *blogsvc.CategoryService, jwt *auth.JWTService) *CategoryHandler {
	return &CategoryHandler{svc: svc, jwt: jwt}
}

func (h *CategoryHandler) Create(ctx context.Context, c *app.RequestContext) {
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

func (h *CategoryHandler) List(ctx context.Context, c *app.RequestContext) {
	title := string(c.Query("title"))
	countPublished := strings.EqualFold(string(c.Query("isDelete")), "true")
	data, err := h.svc.List(ctx, title, countPublished)
	handleAdminResult(ctx, c, data, err)
}

func (h *CategoryHandler) Get(ctx context.Context, c *app.RequestContext) {
	data, err := h.svc.Get(ctx, c.Param("id"))
	handleAdminResult(ctx, c, data, err)
}

func (h *CategoryHandler) Update(ctx context.Context, c *app.RequestContext) {
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

func (h *CategoryHandler) Delete(ctx context.Context, c *app.RequestContext) {
	uid := articleUID(ctx, c, h.jwt)
	if uid == 0 {
		response.Error(ctx, c, errcode.WithMessage(errcode.Unauthorized, "身份验证失败"))
		return
	}
	err := h.svc.Delete(ctx, c.Param("id"))
	handleAdminResult(ctx, c, true, err)
}
