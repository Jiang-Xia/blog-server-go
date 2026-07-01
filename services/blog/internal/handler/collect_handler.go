package handler

import (
	"context"
	"strconv"

	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/pkg/response"
	blogrepo "github.com/Jiang-Xia/blog-server-go/services/blog/internal/blog/repo"
	blogsvc "github.com/Jiang-Xia/blog-server-go/services/blog/internal/blog/service"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/auth"
	"github.com/cloudwego/hertz/pkg/app"
)

// CollectHandler 收藏 HTTP 端点。
type CollectHandler struct {
	svc *blogsvc.CollectService
	jwt *auth.JWTService
}

// NewCollectHandler 构造 CollectHandler。
func NewCollectHandler(svc *blogsvc.CollectService, jwt *auth.JWTService) *CollectHandler {
	return &CollectHandler{svc: svc, jwt: jwt}
}

func (h *CollectHandler) Toggle(ctx context.Context, c *app.RequestContext) {
	uid := articleUID(ctx, c, h.jwt)
	if uid == 0 {
		response.Error(ctx, c, errcode.WithMessage(errcode.Unauthorized, "请先登录！"))
		return
	}
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	articleID, err := blogrepo.ParseCollectArticleID(body["articleId"])
	if err != nil || articleID <= 0 {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	data, err := h.svc.ToggleCollect(ctx, articleID, uid)
	handleAdminResult(ctx, c, data, err)
}

func (h *CollectHandler) Delete(ctx context.Context, c *app.RequestContext) {
	id := c.Param("id")
	if id == "" {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	data, err := h.svc.CancelCollect(ctx, id)
	handleAdminResult(ctx, c, data, err)
}

func (h *CollectHandler) List(ctx context.Context, c *app.RequestContext) {
	uid := articleUID(ctx, c, h.jwt)
	if uid == 0 {
		response.Error(ctx, c, errcode.WithMessage(errcode.Unauthorized, "请先登录！"))
		return
	}
	page, _ := strconv.Atoi(string(c.Query("page")))
	pageSize, _ := strconv.Atoi(string(c.Query("pageSize")))
	data, err := h.svc.GetCollectList(ctx, uid, page, pageSize)
	handleAdminResult(ctx, c, data, err)
}

func (h *CollectHandler) Check(ctx context.Context, c *app.RequestContext) {
	uid := articleUID(ctx, c, h.jwt)
	articleID, _ := strconv.Atoi(string(c.Query("articleId")))
	data, err := h.svc.CheckCollected(ctx, articleID, uid)
	handleAdminResult(ctx, c, data, err)
}

func (h *CollectHandler) Count(ctx context.Context, c *app.RequestContext) {
	articleID, _ := strconv.Atoi(string(c.Query("articleId")))
	data, err := h.svc.GetCollectCount(ctx, articleID)
	handleAdminResult(ctx, c, data, err)
}
