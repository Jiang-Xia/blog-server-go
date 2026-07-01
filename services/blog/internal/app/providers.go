// Package app wire provider 集合。
package app

import (
	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/redisutil"
	"github.com/Jiang-Xia/blog-server-go/pkg/usersvc"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/auth"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/contentfilter"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/event"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/handler"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/middleware"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/userport"
	"github.com/Jiang-Xia/blog-server-go/services/blog/ent"
	bloggrpc "github.com/Jiang-Xia/blog-server-go/services/blog/internal/blog/grpcserver"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/blog/repo"
	blogsvc "github.com/Jiang-Xia/blog-server-go/services/blog/internal/blog/service"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/blog/ws"
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

func provideBlogGRPCServer(articles *blogsvc.ArticleService, client *ent.Client) *bloggrpc.Server {
	return bloggrpc.New(articles, client)
}

func provideBlogEventConsumer(rds rueidis.Client, log *zap.Logger) BlogEventConsumer {
	c := event.NewConsumer(rds, log, event.ConsumerGroupBlog)
	event.RegisterBlogHandlers(c, log)
	return BlogEventConsumer{c}
}

func provideRealtimeRuntime(
	hub *ws.Hub,
	pusher *ws.RealtimePusher,
	blogConsumer BlogEventConsumer,
	publisher *event.Publisher,
	wsH *handler.WSHandler,
	dev *handler.DevPushHandler,
) *RealtimeRuntime {
	return &RealtimeRuntime{
		Hub: hub, Pusher: pusher,
		BlogConsumer: blogConsumer,
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
		WS:           wsH,
		DevPush:      devPushH,
		JWT:          jwt,
		Users:        users,
		Permission: middleware.PermissionDeps{
			Cfg: cfg, Redis: redis, JWT: jwt, Log: log,
		},
	}
}
