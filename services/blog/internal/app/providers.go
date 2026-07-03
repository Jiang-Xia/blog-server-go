// Package app wire provider 集合。
package app

import (
	"fmt"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/redisutil"
	"github.com/Jiang-Xia/blog-server-go/pkg/publicprofile"
	"github.com/Jiang-Xia/blog-server-go/pkg/rpgsvc"
	"github.com/Jiang-Xia/blog-server-go/pkg/usersvc"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/auth"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/contentfilter"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/crossdb"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/event"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/handler"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/middleware"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/rag"
	raglistener "github.com/Jiang-Xia/blog-server-go/services/blog/internal/rag/listener"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/userport"
	"github.com/Jiang-Xia/blog-server-go/services/blog/ent"
	bloggrpc "github.com/Jiang-Xia/blog-server-go/services/blog/internal/blog/grpcserver"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/blog/repo"
	blogsvc "github.com/Jiang-Xia/blog-server-go/services/blog/internal/blog/service"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/blog/scheduler"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/blog/ws"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/scheduledtask"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/scheduledtask/jobs"
	"github.com/redis/rueidis"
	"go.uber.org/zap"
)

func provideArticleRepo(client *ent.Client, cfg *config.Config) (*repo.ArticleRepo, error) {
	return repo.NewArticleRepo(client, cfg)
}

func provideRedisStore(client rueidis.Client) *redisutil.Store {
	return redisutil.New(client)
}

func provideUserService(cfg *config.Config) (usersvc.UserService, error) {
	return userport.ProvideUserService(cfg)
}

func provideBanChecker(cfg *config.Config) (rpgsvc.BanChecker, error) {
	return rpgsvc.NewGRPCBanChecker(cfg.GRPC.RPGAddr)
}

func provideArticleUserPort(users usersvc.UserService) userport.ArticleUserPort {
	return userport.NewGRPCArticleUserPort(users)
}

func provideArticleAdminPort() userport.ArticleAdminPort {
	return userport.NewPermissiveArticleAdminPort()
}

func provideContentFilter() contentfilter.FilterService {
	return contentfilter.NewNoopFilter()
}

func provideJWT(cfg *config.Config) *auth.JWTService {
	return auth.NewJWTService(cfg)
}

func provideBlogGRPCServer(articles *blogsvc.ArticleService, moderation *blogsvc.ModerationService, client *ent.Client, publicProfile *publicprofile.Repo) *bloggrpc.Server {
	return bloggrpc.New(articles, moderation, client, publicProfile)
}

func providePublicProfileRepo(cfg *config.Config) (*publicprofile.Repo, error) {
	return publicprofile.NewRepo(cfg)
}

func provideBlogEventConsumer(rds rueidis.Client, log *zap.Logger) BlogEventConsumer {
	c := event.NewConsumer(rds, log, event.ConsumerGroupBlog)
	event.RegisterBlogHandlers(c, log)
	return BlogEventConsumer{c}
}

func provideRagModule(cfg *config.Config, client *ent.Client, redis *redisutil.Store, articles *repo.ArticleRepo, cross *crossdb.CrossDB, log *zap.Logger) *rag.Module {
	return rag.NewModule(cfg, client, redis, articles, cross, log)
}

func provideRagEventConsumer(rds rueidis.Client, mod *rag.Module, log *zap.Logger) RagEventConsumer {
	if mod == nil || !mod.Cfg.Rag.Enabled {
		return RagEventConsumer{}
	}
	c := event.NewConsumer(rds, log, event.ConsumerGroupRAG)
	raglistener.RegisterRAGHandlers(c, mod, log)
	return RagEventConsumer{c}
}

func provideRealtimeRuntime(
	hub *ws.Hub,
	pusher *ws.RealtimePusher,
	blogConsumer BlogEventConsumer,
	ragConsumer RagEventConsumer,
	publisher *event.Publisher,
	wsH *handler.WSHandler,
	dev *handler.DevPushHandler,
) *RealtimeRuntime {
	return &RealtimeRuntime{
		Hub: hub, Pusher: pusher,
		BlogConsumer: blogConsumer,
		RagConsumer:  ragConsumer,
		Publisher:    publisher,
		WS:           wsH,
		DevPush:      dev,
	}
}

func provideRegisterDeps(
	health *handler.HealthHandler,
	articleH *handler.ArticleHandler,
	categoryH *handler.CategoryHandler,
	tagH *handler.TagHandler,
	commentH *handler.CommentHandler,
	replyH *handler.ReplyHandler,
	likeH *handler.LikeHandler,
	collectH *handler.CollectHandler,
	msgboardH *handler.MsgboardHandler,
	linkH *handler.LinkHandler,
	fileH *handler.FileHandler,
	resourcesH *handler.ResourcesHandler,
	notificationH *handler.NotificationHandler,
	scheduledTaskH *handler.ScheduledTaskHandler,
	ragH *handler.RagHandler,
	wsH *handler.WSHandler,
	devPushH *handler.DevPushHandler,
	jwt *auth.JWTService,
	users usersvc.UserService,
	cfg *config.Config,
	redis *redisutil.Store,
	log *zap.Logger,
) handler.RegisterDeps {
	return handler.RegisterDeps{
		Health:       health,
		Article:      articleH,
		Category:     categoryH,
		Tag:          tagH,
		Comment:      commentH,
		Reply:        replyH,
		Like:         likeH,
		Collect:      collectH,
		Msgboard:     msgboardH,
		Link:         linkH,
		File:         fileH,
		Resources:    resourcesH,
		Notification: notificationH,
		ScheduledTask: scheduledTaskH,
		Rag:           ragH,
		WS:           wsH,
		DevPush:      devPushH,
		JWT:          jwt,
		Users:        users,
		Permission: middleware.PermissionDeps{
			Cfg: cfg, Redis: redis, JWT: jwt, Log: log,
		},
	}
}

func provideScheduledTaskRepo(client *ent.Client) *scheduledtask.Repo {
	return scheduledtask.NewRepo(client)
}

func provideScheduledTaskCrossDB(cfg *config.Config) (*crossdb.CrossDB, error) {
	return crossdb.New(cfg)
}

func provideSystemEmailSender(cfg *config.Config) (usersvc.SystemEmailSender, error) {
	addr := cfg.GRPC.UserAddr
	if addr == "" {
		return nil, fmt.Errorf("GRPC.UserAddr required for system email")
	}
	return usersvc.NewGRPCSystemEmailSender(addr)
}

func provideScheduledTaskJobs(
	client *ent.Client,
	cfg *config.Config,
	articles *repo.ArticleRepo,
	links *repo.LinkRepo,
	repo *scheduledtask.Repo,
	cross *crossdb.CrossDB,
	email usersvc.SystemEmailSender,
	publisher *event.Publisher,
) *jobs.Runner {
	return jobs.NewRunner(client, cfg, articles, links, cross, email, publisher)
}

func provideScheduledTaskService(
	repo *scheduledtask.Repo,
	cross *crossdb.CrossDB,
	cfg *config.Config,
	redis *redisutil.Store,
	runner *jobs.Runner,
	log *zap.Logger,
) *scheduledtask.Service {
	return scheduledtask.NewService(repo, cross, cfg, redis, runner, log)
}

// ScheduledTaskRuntime 连接 cron 与 Service，启动时 Bootstrap。
type ScheduledTaskRuntime struct {
	Sched *scheduler.Scheduler
	Svc   *scheduledtask.Service
}

func provideScheduledTaskRuntime(
	sched *scheduler.Scheduler,
	svc *scheduledtask.Service,
) *ScheduledTaskRuntime {
	sched.SetTrigger(svc.TriggerTask)
	svc.SetScheduler(sched)
	return &ScheduledTaskRuntime{Sched: sched, Svc: svc}
}
