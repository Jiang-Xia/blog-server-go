// Package app user-service 装配与生命周期管理。
package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/services/user/ent"
	"github.com/Jiang-Xia/blog-server-go/services/user/internal/data"
	"github.com/Jiang-Xia/blog-server-go/services/user/internal/user/kitexserver"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/redis/rueidis"
	"go.uber.org/zap"
)

// App 聚合 HTTP/Kitex 服务与基础设施。
type App struct {
	cfg       *config.Config
	h         *server.Hertz
	log       *zap.Logger
	ent       *ent.Client
	redis     rueidis.Client
	userKitex *kitexserver.Server
	cancel    context.CancelFunc
	kitexStop func()
}

// NewApp 构造 user-service 实例。
func NewApp(
	cfg *config.Config,
	h *server.Hertz,
	log *zap.Logger,
	entClient *ent.Client,
	redisClient rueidis.Client,
	userKitex *kitexserver.Server,
) *App {
	return &App{
		cfg: cfg, h: h, log: log, ent: entClient, redis: redisClient,
		userKitex: userKitex,
	}
}

// Run 启动 HTTP/Kitex 并监听退出信号。
func (a *App) Run() error {
	_, cancel := context.WithCancel(context.Background())
	a.cancel = cancel

	if a.userKitex != nil {
		svr, err := kitexserver.Run(a.cfg, a.userKitex)
		if err != nil {
			return fmt.Errorf("start user kitex: %w", err)
		}
		if svr != nil {
			a.log.Info("kitex server started", zap.String("addr", a.cfg.Kitex.Addr))
			a.kitexStop = func() { _ = svr.Stop() }
		}
	}

	go func() {
		a.log.Info("user-service hertz starting")
		a.h.Spin()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	a.log.Info("shutdown signal received", zap.String("signal", sig.String()))
	return a.Shutdown()
}

// Shutdown 优雅关闭。
func (a *App) Shutdown() error {
	if a.cancel != nil {
		a.cancel()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := a.h.Shutdown(ctx); err != nil {
		a.log.Warn("hertz shutdown", zap.Error(err))
	}
	if a.kitexStop != nil {
		a.kitexStop()
	}
	data.CloseEnt(a.ent)
	data.CloseRedis(a.redis)
	_ = a.log.Sync()
	fmt.Println("user-service stopped")
	return nil
}
