// sensitive_handler 敏感词管理 HTTP 端点（管理端）。
package handler

import (
	"context"
	"strconv"
	"strings"

	"github.com/Jiang-Xia/blog-server-go/pkg/ctxutil"
	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/pkg/response"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/auth"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/sensitive"
	"github.com/cloudwego/hertz/pkg/app"
)

// SensitiveWordHandler 敏感词 HTTP 端点。
type SensitiveWordHandler struct {
	svc *sensitive.Service
	jwt *auth.JWTService
}

// NewSensitiveWordHandler 构造 SensitiveWordHandler。
func NewSensitiveWordHandler(svc *sensitive.Service, jwt *auth.JWTService) *SensitiveWordHandler {
	return &SensitiveWordHandler{svc: svc, jwt: jwt}
}

func (h *SensitiveWordHandler) List(ctx context.Context, c *app.RequestContext) {
	page, _ := strconv.Atoi(string(c.Query("page")))
	pageSize, _ := strconv.Atoi(string(c.Query("pageSize")))
	var status *int
	if s := string(c.Query("status")); s != "" {
		v, err := strconv.Atoi(s)
		if err == nil {
			status = &v
		}
	}
	data, err := h.svc.List(ctx, sensitive.ListQuery{
		Page: page, PageSize: pageSize,
		Keyword: string(c.Query("keyword")),
		Category: string(c.Query("category")),
		Status:   status,
	})
	handleAdminResult(ctx, c, data, err)
}

func (h *SensitiveWordHandler) Create(ctx context.Context, c *app.RequestContext) {
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	data, err := h.svc.Create(ctx, body)
	handleAdminResult(ctx, c, data, err)
}

func (h *SensitiveWordHandler) BatchCreate(ctx context.Context, c *app.RequestContext) {
	var raw []map[string]string
	if err := c.Bind(&raw); err != nil {
		var body []map[string]interface{}
		if err2 := c.Bind(&body); err2 != nil {
			response.Error(ctx, c, errcode.InvalidParam)
			return
		}
		raw = make([]map[string]string, 0, len(body))
		for _, item := range body {
			raw = append(raw, map[string]string{
				"word":     strField(item, "word"),
				"category": strField(item, "category"),
			})
		}
	}
	data, err := h.svc.BatchCreate(ctx, raw)
	handleAdminResult(ctx, c, data, err)
}

func (h *SensitiveWordHandler) Update(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	data, err := h.svc.Update(ctx, id, body)
	handleAdminResult(ctx, c, data, err)
}

func (h *SensitiveWordHandler) Delete(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	err = h.svc.Delete(ctx, id)
	handleAdminResult(ctx, c, true, err)
}

func (h *SensitiveWordHandler) ListHits(ctx context.Context, c *app.RequestContext) {
	page, _ := strconv.Atoi(string(c.Query("page")))
	pageSize, _ := strconv.Atoi(string(c.Query("pageSize")))
	data, err := h.svc.ListHits(ctx, sensitive.HitListQuery{
		Page: page, PageSize: pageSize,
		SourceType: string(c.Query("sourceType")),
		Status:     string(c.Query("status")),
	})
	handleAdminResult(ctx, c, data, err)
}

func (h *SensitiveWordHandler) Approve(ctx context.Context, c *app.RequestContext) {
	h.review(ctx, c, true)
}

func (h *SensitiveWordHandler) Reject(ctx context.Context, c *app.RequestContext) {
	h.review(ctx, c, false)
}

func (h *SensitiveWordHandler) review(ctx context.Context, c *app.RequestContext, approve bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	reviewerID := h.uid(ctx, c)
	var data interface{}
	if approve {
		data, err = h.svc.Approve(ctx, id, reviewerID)
	} else {
		data, err = h.svc.Reject(ctx, id, reviewerID)
	}
	handleAdminResult(ctx, c, data, err)
}

func (h *SensitiveWordHandler) uid(ctx context.Context, c *app.RequestContext) int {
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
