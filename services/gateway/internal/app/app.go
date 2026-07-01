// Package app gateway 应用装配与生命周期。
package app

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/logger"
	"github.com/Jiang-Xia/blog-server-go/pkg/metrics"
	"github.com/Jiang-Xia/blog-server-go/pkg/otel"
	"github.com/Jiang-Xia/blog-server-go/pkg/response"
	"github.com/Jiang-Xia/blog-server-go/services/gateway/internal/aggregator"
	gwmw "github.com/Jiang-Xia/blog-server-go/services/gateway/internal/middleware"
	"github.com/Jiang-Xia/blog-server-go/services/gateway/internal/grpcclient"
	"github.com/Jiang-Xia/blog-server-go/services/gateway/internal/proxy"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/adaptor"
	"go.uber.org/zap"
)

// Run 启动 gateway HTTP 服务。
func Run(cfgPath string) error {
	cfg, err := config.MustLoad(cfgPath)
	if err != nil {
		return err
	}
	log, err := logger.New(cfg)
	if err != nil {
		return err
	}

	ctx := context.Background()
	shutdownOTel, err := otel.Init(ctx, cfg)
	if err != nil {
		return err
	}
	defer func() { _ = shutdownOTel(context.Background()) }()

	router, err := proxy.NewRouter(cfg)
	if err != nil {
		return fmt.Errorf("proxy router: %w", err)
	}

	clients, err := grpcclient.New(cfg.GRPC.UserAddr, cfg.GRPC.BlogAddr, cfg.GRPC.RPGAddr)
	if err != nil {
		return fmt.Errorf("grpc clients: %w", err)
	}

	h := server.Default(server.WithHostPorts(cfg.HTTP.Addr))
	h.Use(gwmw.JWTPassthrough(cfg))

	h.GET("/health", func(ctx context.Context, c *app.RequestContext) {
		response.Success(ctx, c, "ok")
	})
	h.GET(cfg.App.APIPrefix+"/health", func(ctx context.Context, c *app.RequestContext) {
		response.Success(ctx, c, "ok")
	})

	stats := aggregator.NewStatsHandler(clients)
	h.GET(cfg.App.APIPrefix+"/pub/stats", stats.Stats)

	article := aggregator.NewArticleHandler(clients)
	h.GET(cfg.App.APIPrefix+"/article/info", article.Info)

	profile := aggregator.NewProfileHandler(clients)
	h.GET(cfg.App.APIPrefix+"/user/public/:uid", profile.PublicProfile)

	if cfg.Observability.EnableMetrics {
		h.GET("/metrics", adaptor.HertzHandler(metrics.Handler()))
	}
	if cfg.Observability.EnablePprof {
		addr := cfg.Observability.PprofAddr
		if addr == "" {
			addr = ":6060"
		}
		go func() {
			_ = http.ListenAndServe(addr, nil)
		}()
		log.Info("pprof listening", zap.String("addr", addr))
	}

	router.Register(h, cfg.App.APIPrefix)

	go func() {
		log.Info("gateway starting", zap.String("addr", cfg.HTTP.Addr))
		h.Spin()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.Info("shutdown", zap.String("signal", sig.String()))
	return h.Shutdown(context.Background())
}
