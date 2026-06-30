// Package app 单体应用装配与生命周期管理（wire 注入入口）。
package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/blog/scheduler"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/blog/ws"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/data"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/redis/rueidis"
	"go.uber.org/zap"
)

// App 聚合 HTTP 服务与基础设施客户端。
type App struct {
	h         *server.Hertz
	log       *zap.Logger
	ent       *ent.Client
	redis     rueidis.Client
	scheduler *scheduler.Scheduler
	realtime  *RealtimeRuntime
	cancel    context.CancelFunc
}

// NewApp 构造可运行的应用实例。
func NewApp(
	h *server.Hertz,
	log *zap.Logger,
	entClient *ent.Client,
	redisClient rueidis.Client,
	sched *scheduler.Scheduler,
	rt *RealtimeRuntime,
) *App {
	return &App{h: h, log: log, ent: entClient, redis: redisClient, scheduler: sched, realtime: rt}
}

// Run 启动 HTTP 服务并监听退出信号。
func (a *App) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel

	if a.realtime != nil && a.realtime.Hub != nil {
		go a.realtime.Hub.Run(ctx)
		ws.StartRedisSubscriber(ctx, a.redis, a.realtime.Hub)
		a.log.Info("websocket hub started")
	}
	if a.realtime != nil && a.realtime.Consumer != nil {
		a.realtime.Consumer.Start(ctx)
		a.log.Info("event consumer started")
	}

	if a.scheduler != nil {
		if err := a.scheduler.RegisterPlaceholder(); err != nil {
			a.log.Warn("register placeholder cron failed", zap.Error(err))
		} else {
			a.scheduler.Start()
			a.log.Info("cron scheduler started")
		}
	}
	go func() {
		a.log.Info("hertz server starting")
		a.h.Spin()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	a.log.Info("shutdown signal received", zap.String("signal", sig.String()))
	return a.Shutdown()
}

// Shutdown 优雅关闭 HTTP、Ent、Redis。
func (a *App) Shutdown() error {
	if a.cancel != nil {
		a.cancel()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := a.h.Shutdown(ctx); err != nil {
		a.log.Warn("hertz shutdown", zap.Error(err))
	}
	if a.scheduler != nil {
		a.scheduler.Stop()
	}
	data.CloseEnt(a.ent)
	data.CloseRedis(a.redis)
	_ = a.log.Sync()
	fmt.Println("server stopped")
	return nil
}
