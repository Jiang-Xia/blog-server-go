// Package server 创建 rpg-service Hertz HTTP 实例。
package server

import (
	"net/http"
	_ "net/http/pprof"

	"github.com/Jiang-Xia/blog-server-go/pkg/apidoc"
	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/metrics"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/internal/handler"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/internal/middleware"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/adaptor"
	"go.uber.org/zap"
)

// NewHTTPServer 装配中间件并注册 RPG/支付域路由。
func NewHTTPServer(cfg *config.Config, log *zap.Logger, deps handler.RegisterDeps) *server.Hertz {
	h := server.Default(server.WithHostPorts(cfg.HTTP.Addr))
	h.NoHijackConnPool = true
	h.Use(
		middleware.Recovery(log),
		middleware.RequestID(),
		middleware.Logger(log),
		middleware.CORS(cfg),
	)
	if cfg.Storage.UploadPath != "" {
		h.Static(cfg.Storage.PublicPrefixOrDefault(), cfg.Storage.UploadPath)
	}
	if cfg.Observability.EnableMetrics {
		h.GET("/metrics", adaptor.HertzHandler(metrics.Handler()))
	}
	if cfg.Observability.EnablePprof {
		addr := cfg.Observability.PprofAddr
		if addr == "" {
			addr = ":6063"
		}
		go func() {
			_ = http.ListenAndServe(addr, nil)
		}()
		log.Info("pprof listening", zap.String("addr", addr))
	}
	handler.Register(h, cfg, deps)
	apidoc.Mount(h, cfg)
	return h
}
