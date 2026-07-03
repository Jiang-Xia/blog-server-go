// Package app wire 依赖注入 provider 集合。
package app

import (
	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/redisutil"
	"github.com/Jiang-Xia/blog-server-go/pkg/rpgsvc"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/blog/operationlog"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/blog/scheduler"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/blog/ws"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/event"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/handler"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/middleware"
	payrepo "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/pay/repo"
	paysvc "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/pay/service"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/pub"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpgport"
	rpgactivity "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/activity"
	rpgevent "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/event"
	"github.com/Jiang-Xia/blog-server-go/pkg/usersvc"
	userpkg "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user"
	usersgrpc "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/grpcserver"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/auth"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/captcha"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/profile"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/repo"
	"github.com/redis/rueidis"
	"go.uber.org/zap"
)

func provideBanChecker(mod *rpg.Module) rpgsvc.BanChecker {
	if mod == nil {
		c, _ := rpgsvc.NewGRPCBanChecker("")
		return c
	}
	return rpgport.NewLocalBanChecker(mod.Punishment)
}

func provideUserGRPCServer(cfg *config.Config, profileSvc *profile.Service, jwt *auth.JWTService) *usersgrpc.Server {
	if cfg.App.ServiceModeOrDefault() != config.ModeUser {
		return nil
	}
	return usersgrpc.New(profileSvc, jwt)
}

func provideUserServicePort(cfg *config.Config, profileSvc *profile.Service) (usersvc.UserService, error) {
	mode := cfg.App.ServiceModeOrDefault()
	if cfg.GRPC.UserAddr != "" && mode != config.ModeMonolith && mode != config.ModeUser {
		return usersvc.NewGRPCUserService(cfg.GRPC.UserAddr)
	}
	return userpkg.NewUserService(profileSvc), nil
}

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

func providePayOrderRepo(client *ent.Client) *payrepo.PayOrderRepo {
	return payrepo.NewPayOrderRepo(client)
}

func providePayService(cfg *config.Config, orderRepo *payrepo.PayOrderRepo, log *zap.Logger) (*paysvc.PayService, error) {
	return paysvc.NewPayService(cfg, orderRepo, log)
}

func providePayOrderService(orderRepo *payrepo.PayOrderRepo, pay *paysvc.PayService, mod *rpg.Module, log *zap.Logger) *paysvc.PayOrderService {
	return providePayOrderServiceWithRecharge(orderRepo, pay, mod, log)
}

func provideRPGHandler(mod *rpg.Module, game *handler.RPGGameplay, jwt *auth.JWTService) *handler.RPGHandler {
	return handler.NewRPGHandler(mod, game, jwt)
}

func provideActivityNotifyScheduler(mod *rpg.Module, log *zap.Logger) scheduler.ActivityNotifyRunner {
	return rpgactivity.NewNotifyScheduler(mod.Repo, mod.Notify, log)
}

// BlogEventConsumer blog 域 Stream 消费器（wire 区分类型用）。
type BlogEventConsumer struct {
	*event.Consumer
}

// RPGEventConsumer RPG 域 Stream 消费器（wire 区分类型用）。
type RPGEventConsumer struct {
	*event.Consumer
}

// RealtimeRuntime Hub 与 Stream 消费者生命周期。
type RealtimeRuntime struct {
	Hub          *ws.Hub
	Pusher       *ws.RealtimePusher
	BlogConsumer BlogEventConsumer
	RPGConsumer  RPGEventConsumer
	Publisher    *event.Publisher
	WS           *handler.WSHandler
	DevPush      *handler.DevPushHandler
}

func provideBlogEventConsumer(rds rueidis.Client, log *zap.Logger) BlogEventConsumer {
	c := event.NewConsumer(rds, log, event.ConsumerGroupBlog)
	event.RegisterBlogHandlers(c, log)
	return BlogEventConsumer{c}
}

func provideRPGEventHandlers(mod *rpg.Module, redis *redisutil.Store) rpgevent.Handlers {
	return provideRPGEventHandlersFull(mod, redis)
}

func provideRPGEventConsumer(rds rueidis.Client, log *zap.Logger, handlers rpgevent.Handlers) RPGEventConsumer {
	c := event.NewConsumer(rds, log, event.ConsumerGroupRPG)
	rpgevent.RegisterRPGHandlers(c, handlers)
	return RPGEventConsumer{c}
}

func provideRealtimeRuntime(
	hub *ws.Hub,
	pusher *ws.RealtimePusher,
	blogConsumer BlogEventConsumer,
	rpgConsumer RPGEventConsumer,
	publisher *event.Publisher,
	wsH *handler.WSHandler,
	dev *handler.DevPushHandler,
) *RealtimeRuntime {
	return &RealtimeRuntime{
		Hub: hub, Pusher: pusher,
		BlogConsumer: blogConsumer, RPGConsumer: rpgConsumer,
		Publisher: publisher, WS: wsH, DevPush: dev,
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
	rpgH *handler.RPGHandler,
	rpgAdminH *handler.RPGAdminHandler,
	rpgProfileH *handler.RPGProfileHandler,
	payH *handler.PayHandler,
	payOrderH *handler.PayOrderHandler,
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
		RPG:          rpgH,
		RPGAdmin:     rpgAdminH,
		RPGProfile:   rpgProfileH,
		Pay:          payH,
		PayOrder:     payOrderH,
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
