//go:generate go run github.com/google/wire/cmd/wire
//go:build wireinject
// +build wireinject

// Package app wire 依赖注入声明（make wire-rpg 生成 wire_gen.go）。
package app

import (
	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/logger"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/internal/event"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/internal/handler"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/internal/scheduler"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/internal/server"
	"github.com/Jiang-Xia/blog-server-go/pkg/usersvc"
	"github.com/google/wire"
)

// InitializeApp 装配 rpg-service 全部依赖。
func InitializeApp(cfgPath string) (*App, error) {
	wire.Build(
		config.MustLoad,
		logger.New,
		provideDataStore,
		provideEntClient,
		provideSQLDB,
		provideRedisClient,
		provideRedisStore,
		provideUserService,
		wire.Bind(new(usersvc.UserService), new(usersvc.CrossClient)),
		wire.Bind(new(usersvc.SensitiveHitLister), new(usersvc.CrossClient)),
		provideUserReader,
		provideArticleReader,
		provideBlogPublicProfileLister,
		provideBlogArticleRPGStore,
		provideWSPusher,
		provideJWT,
		rpg.NewModule,
		event.NewPublisher,
		provideRPGEventHandlers,
		provideRPGEventConsumer,
		provideRPGGameplay,
		provideRPGHandler,
		provideRPGAdminHandler,
		provideRPGProfileHandler,
		providePayOrderRepo,
		providePayService,
		providePayOrderService,
		handler.NewPayHandler,
		handler.NewPayOrderHandler,
		handler.NewHealthHandler,
		scheduler.New,
		provideActivityNotifyScheduler,
		provideRegisterDeps,
		provideRPGGRPCServer,
		server.NewHTTPServer,
		NewApp,
	)
	return nil, nil
}
