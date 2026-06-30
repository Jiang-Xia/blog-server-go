// Package handler 注册 HTTP 路由与 health 等基础端点。
package handler

import (
	"context"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/response"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/redis/rueidis"
)

// HealthHandler 健康检查与连通性探测。
type HealthHandler struct {
	ent   *ent.Client
	redis rueidis.Client
}

// NewHealthHandler 构造 HealthHandler。
func NewHealthHandler(entClient *ent.Client, redisClient rueidis.Client) *HealthHandler {
	return &HealthHandler{ent: entClient, redis: redisClient}
}

// OK 返回统一成功响应；启动时已验证 MySQL/Redis，此处仅回显 ok。
func (h *HealthHandler) OK(ctx context.Context, c *app.RequestContext) {
	response.Success(ctx, c, "ok")
}

// Register 注册 /health 与 /api/v1/health。
func Register(r *server.Hertz, cfg *config.Config, health *HealthHandler) {
	r.GET("/health", health.OK)
	r.GET(cfg.App.APIPrefix+"/health", health.OK)
}
