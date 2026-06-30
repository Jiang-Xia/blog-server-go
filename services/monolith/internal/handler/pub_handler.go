// Package handler 公开统计 HTTP 端点。
package handler

import (
	"context"

	"github.com/Jiang-Xia/blog-server-go/pkg/response"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/pub"
	"github.com/cloudwego/hertz/pkg/app"
)

// PubHandlerDeps pub handler 依赖。
type PubHandlerDeps struct {
	Pub *pub.Service
}

// PubHandler 公开接口 handler。
type PubHandler struct {
	pub *pub.Service
}

// NewPubHandler 构造 PubHandler。
func NewPubHandler(deps PubHandlerDeps) *PubHandler {
	return &PubHandler{pub: deps.Pub}
}

// Stats GET /pub/stats — 站点统计（blog 计数 mock，用户数读库）。
func (h *PubHandler) Stats(ctx context.Context, c *app.RequestContext) {
	stats, err := h.pub.GetStats(ctx)
	if err != nil {
		response.FromError(ctx, c, err)
		return
	}
	response.Success(ctx, c, stats)
}
