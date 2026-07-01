// Package app 单体应用装配与生命周期管理（wire 注入入口）。
package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/blog/scheduler"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/blog/ws"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/data"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg"
	rpgseeds "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/seeds"
	usersgrpc "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/grpcserver"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/redis/rueidis"
	"go.uber.org/zap"
)

// App 聚合 HTTP 服务与基础设施客户端。
type App struct {
	cfg             *config.Config
	h               *server.Hertz
	log             *zap.Logger
	ent             *ent.Client
	redis           rueidis.Client
	scheduler       *scheduler.Scheduler
	realtime        *RealtimeRuntime
	rpgMod          *rpg.Module
	activityNotify  scheduler.ActivityNotifyRunner
	userGRPC        *usersgrpc.Server
	cancel          context.CancelFunc
	grpcStop        func()
}

// NewApp 构造可运行的应用实例。
func NewApp(
	cfg *config.Config,
	h *server.Hertz,
	log *zap.Logger,
	entClient *ent.Client,
	redisClient rueidis.Client,
	sched *scheduler.Scheduler,
	rt *RealtimeRuntime,
	rpgMod *rpg.Module,
	activityNotify scheduler.ActivityNotifyRunner,
	userGRPC *usersgrpc.Server,
) *App {
	return &App{
		cfg: cfg, h: h, log: log, ent: entClient, redis: redisClient,
		scheduler: sched, realtime: rt, rpgMod: rpgMod, activityNotify: activityNotify,
		userGRPC: userGRPC,
	}
}

// Run 启动 HTTP/gRPC 服务并监听退出信号。
func (a *App) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel

	mode := a.cfg.App.ServiceModeOrDefault()

	if a.userGRPC != nil && a.cfg.GRPC.Addr != "" {
		gs, err := usersgrpc.Run(a.cfg.GRPC.Addr, a.userGRPC)
		if err != nil {
			return fmt.Errorf("start user grpc: %w", err)
		}
		if gs != nil {
			a.log.Info("user grpc server started", zap.String("addr", a.cfg.GRPC.Addr))
			a.grpcStop = func() { gs.GracefulStop() }
		}
	}

	if a.rpgMod != nil && a.rpgMod.Repo != nil && (mode == config.ModeMonolith || mode == config.ModeRPG) {
		if err := rpgseeds.SyncAllPredefined(ctx, a.rpgMod.Repo, a.log); err != nil {
			a.log.Warn("rpg predefined seed sync failed", zap.Error(err))
		} else {
			a.log.Info("rpg predefined seeds synced")
		}
	}

	if a.realtime != nil && a.realtime.Hub != nil && (mode == config.ModeMonolith || mode == config.ModeBlog) {
		go a.realtime.Hub.Run(ctx)
		ws.StartRedisSubscriber(ctx, a.redis, a.realtime.Hub)
		a.log.Info("websocket hub started")
	}
	if a.realtime != nil && a.realtime.BlogConsumer.Consumer != nil && (mode == config.ModeMonolith || mode == config.ModeBlog) {
		a.realtime.BlogConsumer.Consumer.Start(ctx)
		a.log.Info("blog event consumer started")
	}
	if a.realtime != nil && a.realtime.RPGConsumer.Consumer != nil && (mode == config.ModeMonolith || mode == config.ModeRPG) {
		a.realtime.RPGConsumer.Consumer.Start(ctx)
		a.log.Info("rpg event consumer started")
	}

	if a.scheduler != nil && mode != config.ModeUser {
		if err := a.scheduler.RegisterPlaceholder(); err != nil {
			a.log.Warn("register placeholder cron failed", zap.Error(err))
		}
		if err := a.scheduler.RegisterActivityNotify(a.activityNotify); err != nil {
			a.log.Warn("register activity notify cron failed", zap.Error(err))
		}
		a.scheduler.Start()
		a.log.Info("cron scheduler started")
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

// Shutdown 优雅关闭 HTTP、gRPC、Ent、Redis。
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
	if a.grpcStop != nil {
		a.grpcStop()
	}
	data.CloseEnt(a.ent)
	data.CloseRedis(a.redis)
	_ = a.log.Sync()
	fmt.Println("server stopped")
	return nil
}
