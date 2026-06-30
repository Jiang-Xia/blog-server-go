// Package app wire 依赖注入 provider 集合。
package app

import (
	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/redisutil"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/blog/operationlog"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/blog/ws"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/event"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/handler"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/middleware"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/pub"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/auth"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/captcha"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/repo"
	"github.com/redis/rueidis"
	"go.uber.org/zap"
)

func provideAdminRepo(cfg *config.Config) (*repo.AdminRepo, error) {
	return repo.NewAdminRepo(cfg)
}

func provideRoleRepo(cfg *config.Config) (*repo.RoleRepo, error) {
	return repo.NewRoleRepo(cfg)
}

func provideRedisStore(client rueidis.Client) *redisutil.Store {
	return redisutil.New(client)
}

func providePasswordChecker(cfg *config.Config, userRepo *repo.UserRepo) *auth.PasswordChecker {
	return auth.NewPasswordChecker(userRepo, cfg.Crypto.RSAPrivateKeyOrDefault())
}

func provideUserHandler(cfg *config.Config, appSvc *handler.UserAppAdapter, captchaSvc *captcha.Service) *handler.UserHandler {
	return handler.NewUserHandler(handler.UserHandlerDeps{Cfg: cfg, Svc: appSvc, Captcha: captchaSvc})
}

func provideCaptchaHandler(cfg *config.Config, captchaSvc *captcha.Service) *handler.CaptchaHandler {
	return handler.NewCaptchaHandler(handler.CaptchaHandlerDeps{Cfg: cfg, Captcha: captchaSvc})
}

func providePubHandler(pubSvc *pub.Service) *handler.PubHandler {
	return handler.NewPubHandler(handler.PubHandlerDeps{Pub: pubSvc})
}

// RealtimeRuntime Hub 与 Stream 消费者生命周期。
type RealtimeRuntime struct {
	Hub       *ws.Hub
	Pusher    *ws.RealtimePusher
	Consumer  *event.Consumer
	Publisher *event.Publisher
	WS        *handler.WSHandler
	DevPush   *handler.DevPushHandler
}

func provideEventConsumer(rds rueidis.Client, log *zap.Logger) *event.Consumer {
	c := event.NewConsumer(rds, log, event.ConsumerGroupBlog)
	event.RegisterBlogHandlers(c, log)
	return c
}

func provideRealtimeRuntime(
	hub *ws.Hub,
	pusher *ws.RealtimePusher,
	consumer *event.Consumer,
	publisher *event.Publisher,
	wsH *handler.WSHandler,
	dev *handler.DevPushHandler,
) *RealtimeRuntime {
	return &RealtimeRuntime{
		Hub: hub, Pusher: pusher, Consumer: consumer, Publisher: publisher, WS: wsH, DevPush: dev,
	}
}

func provideRegisterDeps(
	health *handler.HealthHandler,
	userH *handler.UserHandler,
	adminH *handler.AdminHandler,
	captchaH *handler.CaptchaHandler,
	pubH *handler.PubHandler,
	sensitiveH *handler.SensitiveWordHandler,
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
	operationLogH *handler.OperationLogHandler,
	wsH *handler.WSHandler,
	devPushH *handler.DevPushHandler,
	jwt *auth.JWTService,
	userRepo *repo.UserRepo,
	cfg *config.Config,
	redis *redisutil.Store,
	roleRepo *repo.RoleRepo,
	opLogSvc *operationlog.Service,
	log *zap.Logger,
) handler.RegisterDeps {
	return handler.RegisterDeps{
		Health:       health,
		User:         userH,
		Admin:        adminH,
		Captcha:      captchaH,
		Pub:          pubH,
		Sensitive:    sensitiveH,
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
		OperationLog: operationLogH,
		WS:           wsH,
		DevPush:      devPushH,
		JWT:          jwt,
		UserRepo:     userRepo,
		Permission: middleware.PermissionDeps{
			Cfg: cfg, Redis: redis, RoleRepo: roleRepo, JWT: jwt, Log: log,
		},
		OpLog: middleware.OperationLogDeps{
			Svc: opLogSvc, JWT: jwt, Log: log,
		},
	}
}
