// Package app blog-service 装配与生命周期管理。
package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/services/blog/ent"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/blog/kitexserver"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/blog/ws"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/data"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/event"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/handler"
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
	schedTask *ScheduledTaskRuntime
	realtime  *RealtimeRuntime
	blogKitex *kitexserver.Server
	cancel    context.CancelFunc
	kitexStop func()
}

// NewApp 构造 blog-service 实例。
func NewApp(
	cfg *config.Config,
	h *server.Hertz,
	log *zap.Logger,
	entClient *ent.Client,
	redisClient rueidis.Client,
	schedTask *ScheduledTaskRuntime,
	rt *RealtimeRuntime,
	blogKitex *kitexserver.Server,
) *App {
	return &App{
		cfg: cfg, h: h, log: log, ent: entClient, redis: redisClient,
		schedTask: schedTask, realtime: rt, blogKitex: blogKitex,
	}
}

// Run 启动 HTTP、WS Hub 与事件消费并监听退出信号。
func (a *App) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel

	if a.blogKitex != nil {
		svr, err := kitexserver.Run(a.cfg, a.blogKitex)
		if err != nil {
			return fmt.Errorf("start blog kitex: %w", err)
		}
		if svr != nil {
			a.log.Info("kitex server started", zap.String("addr", a.cfg.Kitex.Addr))
			a.kitexStop = func() { _ = svr.Stop() }
		}
	}

	if a.realtime != nil && a.realtime.Hub != nil {
		go a.realtime.Hub.Run(ctx)
		ws.StartRedisSubscriber(ctx, a.redis, a.realtime.Hub)
		a.log.Info("websocket hub started")
	}
	if a.realtime != nil && a.realtime.BlogConsumer.Consumer != nil {
		a.realtime.BlogConsumer.Consumer.Start(ctx)
		a.log.Info("blog event consumer started")
	}
	if a.realtime != nil && a.realtime.RagConsumer.Consumer != nil {
		a.realtime.RagConsumer.Consumer.Start(ctx)
		a.log.Info("rag event consumer started")
	}

	if a.schedTask != nil && a.schedTask.Sched != nil {
		if err := a.schedTask.Svc.Bootstrap(ctx); err != nil {
			a.log.Warn("scheduled task bootstrap failed", zap.Error(err))
		}
		a.schedTask.Sched.Start()
		a.log.Info("cron scheduler started")
	}

	go func() {
		a.log.Info("blog-service hertz starting")
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
	if a.schedTask != nil && a.schedTask.Sched != nil {
		a.schedTask.Sched.Stop()
	}
	if a.kitexStop != nil {
		a.kitexStop()
	}
	data.CloseEnt(a.ent)
	data.CloseRedis(a.redis)
	_ = a.log.Sync()
	fmt.Println("blog-service stopped")
	return nil
}

// BlogEventConsumer blog 域 Stream 消费器（wire 区分类型用）。
type BlogEventConsumer struct {
	*event.Consumer
}

// RagEventConsumer RAG 增量索引 Stream 消费器。
type RagEventConsumer struct {
	*event.Consumer
}

// RealtimeRuntime Hub 与 Stream 消费者生命周期。
type RealtimeRuntime struct {
	Hub          *ws.Hub
	Pusher       *ws.RealtimePusher
	BlogConsumer BlogEventConsumer
	RagConsumer  RagEventConsumer
	Publisher    *event.Publisher
	WS           *handler.WSHandler
	DevPush      *handler.DevPushHandler
}
