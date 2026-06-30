// Package server 创建并配置 Hertz HTTP 服务实例。
package server

import (
	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/handler"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/middleware"
	"github.com/cloudwego/hertz/pkg/app/server"
	"go.uber.org/zap"
)

// NewHTTPServer 装配中间件并注册基础路由。
func NewHTTPServer(cfg *config.Config, log *zap.Logger, health *handler.HealthHandler) *server.Hertz {
	h := server.Default(server.WithHostPorts(cfg.HTTP.Addr))
	h.Use(
		middleware.Recovery(log),
		middleware.RequestID(),
		middleware.Logger(log),
		middleware.CORS(cfg),
	)
	handler.Register(h, cfg, health)
	return h
}
