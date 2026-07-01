// Package app rpg-service 装配与生命周期管理。
package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/internal/data"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg"
	rpggrpc "github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/grpcserver"
	rpgseeds "github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/seeds"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/internal/scheduler"
	"github.com/cloudwego/hertz/pkg/app/server"
	"go.uber.org/zap"
)

// App 聚合 HTTP 服务与基础设施。
type App struct {
	cfg            *config.Config
	h              *server.Hertz
	log            *zap.Logger
	data           *data.Data
	scheduler      *scheduler.Scheduler
	activityNotify scheduler.ActivityNotifyRunner
	rpgMod         *rpg.Module
	rpgConsumer    RPGEventConsumer
	rpgGRPC        *rpggrpc.Server
	cancel         context.CancelFunc
	grpcStop       func()
}

// NewApp 构造 rpg-service 实例。
func NewApp(
	cfg *config.Config,
	h *server.Hertz,
	log *zap.Logger,
	d *data.Data,
	sched *scheduler.Scheduler,
	activityNotify scheduler.ActivityNotifyRunner,
	rpgMod *rpg.Module,
	rpgConsumer RPGEventConsumer,
	rpgGRPC *rpggrpc.Server,
) *App {
	return &App{
		cfg: cfg, h: h, log: log, data: d,
		scheduler: sched, activityNotify: activityNotify, rpgMod: rpgMod,
		rpgConsumer: rpgConsumer, rpgGRPC: rpgGRPC,
	}
}

// Run 启动 HTTP、事件消费与 cron。
func (a *App) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel

	if a.rpgGRPC != nil && a.cfg.GRPC.Addr != "" {
		gs, err := rpggrpc.Run(a.cfg.GRPC.Addr, a.rpgGRPC)
		if err != nil {
			return fmt.Errorf("start rpg grpc: %w", err)
		}
		if gs != nil {
			a.log.Info("rpg grpc server started", zap.String("addr", a.cfg.GRPC.Addr))
			a.grpcStop = func() { gs.GracefulStop() }
		}
	}

	if a.rpgMod != nil && a.rpgMod.Repo != nil {
		if err := rpgseeds.SyncAllPredefined(ctx, a.rpgMod.Repo, a.log); err != nil {
			a.log.Warn("rpg predefined seed sync failed", zap.Error(err))
		} else {
			a.log.Info("rpg predefined seeds synced")
		}
	}

	if a.rpgConsumer.Consumer != nil {
		a.rpgConsumer.Consumer.Start(ctx)
		a.log.Info("rpg event consumer started")
	}

	if a.scheduler != nil {
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
		a.log.Info("rpg-service hertz starting")
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
	if a.scheduler != nil {
		a.scheduler.Stop()
	}
	if a.grpcStop != nil {
		a.grpcStop()
	}
	if a.data != nil {
		a.data.Close()
	}
	_ = a.log.Sync()
	fmt.Println("rpg-service stopped")
	return nil
}
