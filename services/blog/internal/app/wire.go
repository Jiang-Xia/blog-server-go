//go:generate go run github.com/google/wire/cmd/wire
//go:build wireinject
// +build wireinject

// Package app wire 依赖注入声明（make wire-blog 生成 wire_gen.go）。
package app

import (
	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/logger"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/blog/notification"
	blogrepo "github.com/Jiang-Xia/blog-server-go/services/blog/internal/blog/repo"
	blogsvc "github.com/Jiang-Xia/blog-server-go/services/blog/internal/blog/service"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/blog/scheduler"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/blog/ws"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/data"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/event"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/handler"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/server"
	"github.com/google/wire"
)

// InitializeApp 装配 blog-service 全部依赖。
func InitializeApp(cfgPath string) (*App, error) {
	wire.Build(
		config.MustLoad,
		logger.New,
		data.NewEntClient,
		data.NewRedisClient,
		provideRedisStore,
		provideUserService,
		provideBanChecker,
		provideArticleUserPort,
		provideArticleAdminPort,
		provideContentFilter,
		provideJWT,
		provideArticleRepo,
		blogrepo.NewCategoryRepo,
		blogrepo.NewTagRepo,
		blogrepo.NewCommentRepo,
		blogrepo.NewReplyRepo,
		blogrepo.NewLikeRepo,
		blogrepo.NewCollectRepo,
		blogrepo.NewMsgboardRepo,
		blogrepo.NewLinkRepo,
		blogrepo.NewFileRepo,
		blogsvc.NewCategoryService,
		blogsvc.NewTagService,
		blogsvc.NewReplyService,
		blogsvc.NewCommentService,
		blogsvc.NewLikeService,
		blogsvc.NewCollectService,
		blogsvc.NewMsgboardService,
		blogsvc.NewModerationService,
		blogsvc.NewLinkService,
		blogsvc.NewResourcesService,
		blogsvc.NewArticleService,
		ws.NewHub,
		ws.NewRealtimePusher,
		wire.Bind(new(ws.Pusher), new(*ws.RealtimePusher)),
		event.NewPublisher,
		provideBlogEventConsumer,
		provideRealtimeRuntime,
		notification.NewService,
		scheduler.New,
		provideScheduledTaskRepo,
		provideScheduledTaskCrossDB,
		provideSystemEmailSender,
		provideScheduledTaskJobs,
		provideScheduledTaskService,
		provideScheduledTaskRuntime,
		handler.NewScheduledTaskHandler,
		handler.NewArticleHandler,
		handler.NewCategoryHandler,
		handler.NewTagHandler,
		handler.NewCommentHandler,
		handler.NewReplyHandler,
		handler.NewLikeHandler,
		handler.NewCollectHandler,
		handler.NewMsgboardHandler,
		handler.NewLinkHandler,
		handler.NewFileHandler,
		handler.NewResourcesHandler,
		handler.NewNotificationHandler,
		handler.NewWSHandler,
		handler.NewDevPushHandler,
		handler.NewHealthHandler,
		provideRegisterDeps,
		provideBlogGRPCServer,
		server.NewHTTPServer,
		NewApp,
	)
	return nil, nil
}
