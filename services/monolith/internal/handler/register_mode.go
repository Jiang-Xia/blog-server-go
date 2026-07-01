// Package handler 按 ServiceMode 注册路由。
package handler

import (
	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/cloudwego/hertz/pkg/app/server"
)

// RegisterByMode 按运行形态注册路由；monolith 注册全部。
func RegisterByMode(r *server.Hertz, cfg *config.Config, deps RegisterDeps) {
	switch cfg.App.ServiceModeOrDefault() {
	case config.ModeUser:
		r.GET("/health", deps.Health.OK)
		r.GET(cfg.App.APIPrefix+"/health", deps.Health.OK)
		RegisterUser(r, cfg, deps)
	case config.ModeBlog:
		r.GET("/health", deps.Health.OK)
		r.GET(cfg.App.APIPrefix+"/health", deps.Health.OK)
		RegisterBlog(r, cfg, deps)
	case config.ModeRPG:
		r.GET("/health", deps.Health.OK)
		r.GET(cfg.App.APIPrefix+"/health", deps.Health.OK)
		RegisterRPG(r, cfg, deps)
	default:
		RegisterAll(r, cfg, deps)
	}
}
