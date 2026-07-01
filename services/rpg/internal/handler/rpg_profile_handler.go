// Package handler 公开主页与 RPG 展示 HTTP 端点，路径对齐 Nest ProfileController。
package handler

import (
	"context"
	"strconv"
	"strings"

	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/pkg/response"
	"github.com/cloudwego/hertz/pkg/app"
)

// RPGProfileService 公开主页 RPG 数据（由 rpg/profile 实现）。
type RPGProfileService interface {
	GetPublicProfile(ctx context.Context, uid int) (interface{}, error)
	GetPublicArticles(ctx context.Context, uid, page, pageSize int) (interface{}, error)
	GetPublicCollectArticles(ctx context.Context, uid, page, pageSize int) (interface{}, error)
	GetPublicLikeArticles(ctx context.Context, uid, page, pageSize int) (interface{}, error)
	ParsePublicRpgUIDs(raw string) []int
	GetPublicRpgLevelsBatch(ctx context.Context, uids []int) (interface{}, error)
	GetPublicRpgStatus(ctx context.Context, uid int) (interface{}, error)
}

// RPGProfileHandler 公开 user/public 与 rpg/public 路由。
type RPGProfileHandler struct {
	svc RPGProfileService
}

// NewRPGProfileHandler 构造 RPGProfileHandler。
func NewRPGProfileHandler(svc RPGProfileService) *RPGProfileHandler {
	return &RPGProfileHandler{svc: svc}
}

func (h *RPGProfileHandler) requireSvc(ctx context.Context, c *app.RequestContext) bool {
	if h.svc == nil {
		response.Error(ctx, c, errcode.WithMessage(errcode.InternalError, "公开主页模块加载中"))
		return false
	}
	return true
}

func (h *RPGProfileHandler) PublicProfile(ctx context.Context, c *app.RequestContext) {
	if !h.requireSvc(ctx, c) {
		return
	}
	uid, _ := strconv.Atoi(c.Param("uid"))
	data, err := h.svc.GetPublicProfile(ctx, uid)
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGProfileHandler) PublicArticles(ctx context.Context, c *app.RequestContext) {
	if !h.requireSvc(ctx, c) {
		return
	}
	uid, _ := strconv.Atoi(c.Param("uid"))
	data, err := h.svc.GetPublicArticles(ctx, uid, queryInt(c, "page", 1), queryInt(c, "pageSize", 10))
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGProfileHandler) PublicCollects(ctx context.Context, c *app.RequestContext) {
	if !h.requireSvc(ctx, c) {
		return
	}
	uid, _ := strconv.Atoi(c.Param("uid"))
	data, err := h.svc.GetPublicCollectArticles(ctx, uid, queryInt(c, "page", 1), queryInt(c, "pageSize", 10))
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGProfileHandler) PublicLikes(ctx context.Context, c *app.RequestContext) {
	if !h.requireSvc(ctx, c) {
		return
	}
	uid, _ := strconv.Atoi(c.Param("uid"))
	data, err := h.svc.GetPublicLikeArticles(ctx, uid, queryInt(c, "page", 1), queryInt(c, "pageSize", 10))
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGProfileHandler) PublicRpgStatusBatch(ctx context.Context, c *app.RequestContext) {
	if !h.requireSvc(ctx, c) {
		return
	}
	uids := h.svc.ParsePublicRpgUIDs(string(c.Query("uids")))
	data, err := h.svc.GetPublicRpgLevelsBatch(ctx, uids)
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGProfileHandler) PublicRpgStatus(ctx context.Context, c *app.RequestContext) {
	if !h.requireSvc(ctx, c) {
		return
	}
	uid, _ := strconv.Atoi(c.Param("uid"))
	data, err := h.svc.GetPublicRpgStatus(ctx, uid)
	handleAdminResult(ctx, c, data, err)
}

// ParseCommaUIDs 解析逗号分隔 uid 列表（备用）。
func ParseCommaUIDs(raw string) []int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		if n, err := strconv.Atoi(strings.TrimSpace(p)); err == nil {
			out = append(out, n)
		}
	}
	return out
}
