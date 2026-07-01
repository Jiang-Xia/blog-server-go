// Package handler 健康检查端点。
package handler

import (
	"context"

	"github.com/Jiang-Xia/blog-server-go/pkg/response"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/ent"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/redis/rueidis"
)

// HealthHandler 健康检查。
type HealthHandler struct {
	ent   *ent.Client
	redis rueidis.Client
}

// NewHealthHandler 构造 HealthHandler。
func NewHealthHandler(entClient *ent.Client, redisClient rueidis.Client) *HealthHandler {
	return &HealthHandler{ent: entClient, redis: redisClient}
}

// OK 返回统一成功响应。
func (h *HealthHandler) OK(ctx context.Context, c *app.RequestContext) {
	response.Success(ctx, c, "ok")
}
