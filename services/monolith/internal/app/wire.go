//go:generate go run github.com/google/wire/cmd/wire
//go:build wireinject
// +build wireinject

// Package app wire 依赖注入声明（make wire 生成 wire_gen.go）。
package app

import (
	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/logger"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/blog/notification"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/blog/operationlog"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/blog/ws"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/event"
	blogrepo "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/blog/repo"
	blogsvc "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/blog/service"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/blog/scheduler"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/data"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/handler"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/pub"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/server"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/admin"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/sensitive"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/auth"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/captcha"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/email"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/profile"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/repo"
	"github.com/google/wire"
)

// InitializeApp 装配单体应用全部依赖。
func InitializeApp(cfgPath string) (*App, error) {
	wire.Build(
		config.MustLoad,
		logger.New,
		data.NewEntClient,
		data.NewRedisClient,
		provideRedisStore,
		repo.NewUserRepo,
		provideRoleRepo,
		provideAdminRepo,
		auth.NewJWTService,
		providePasswordChecker,
		email.NewService,
		auth.NewAuthService,
		auth.NewGitHubOAuth,
		profile.NewService,
		admin.NewService,
		sensitive.NewService,
		wire.Bind(new(sensitive.FilterService), new(*sensitive.Service)),
		blogrepo.NewArticleRepo,
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
		blogsvc.NewLinkService,
		blogsvc.NewResourcesService,
		provideUserServicePort,
		blogsvc.NewArticleService,
		ws.NewHub,
		ws.NewRealtimePusher,
		wire.Bind(new(ws.Pusher), new(*ws.RealtimePusher)),
		event.NewPublisher,
		provideBlogEventConsumer,
		provideRPGEventConsumer,
		provideCrossDB,
		provideRagModule,
		provideRagEventConsumer,
		notification.NewService,
		operationlog.NewService,
		scheduler.New,
		provideScheduledTaskRepo,
		provideScheduledTaskJobs,
		provideScheduledTaskService,
		provideScheduledTaskRuntime,
		rpg.NewModule,
		provideBanChecker,
		provideRPGEventHandlers,
		providePayOrderRepo,
		providePayService,
		providePayOrderService,
		provideRPGGameplay,
		providePublicProfileRepo,
		captcha.NewService,
		pub.NewService,
		handler.NewUserAppAdapter,
		provideUserHandler,
		handler.NewAdminHandler,
		handler.NewSensitiveWordHandler,
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
		handler.NewScheduledTaskHandler,
		handler.NewRagHandler,
		handler.NewWSHandler,
		handler.NewDevPushHandler,
		handler.NewOperationLogHandler,
		provideRPGHandler,
		provideRPGAdminHandler,
		provideRPGProfileHandler,
		handler.NewPayHandler,
		handler.NewPayOrderHandler,
		provideActivityNotifyScheduler,
		provideRealtimeRuntime,
		provideCaptchaHandler,
		providePubHandler,
		handler.NewHealthHandler,
		provideRegisterDeps,
		server.NewHTTPServer,
		NewApp,
	)
	return nil, nil
}
