// Package handler 博客域 HTTP 与 WebSocket 路由注册。
//
// 入口：/health 无鉴权；业务路由见 RegisterBlog。
package handler

import (
	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/cloudwego/hertz/pkg/app/server"
)

// Register 注册 health 与博客域路由。
func Register(r *server.Hertz, cfg *config.Config, deps RegisterDeps) {
	r.GET("/health", deps.Health.OK)
	r.GET(cfg.App.APIPrefix+"/health", deps.Health.OK)
	RegisterBlog(r, cfg, deps)
}
