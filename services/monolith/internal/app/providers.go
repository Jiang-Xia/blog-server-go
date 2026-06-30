// Package app wire 依赖注入 provider 集合。
package app

import (
	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/redisutil"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/blog/operationlog"
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

func provideRegisterDeps(
	health *handler.HealthHandler,
	userH *handler.UserHandler,
	adminH *handler.AdminHandler,
	captchaH *handler.CaptchaHandler,
	pubH *handler.PubHandler,
	sensitiveH *handler.SensitiveWordHandler,
	notificationH *handler.NotificationHandler,
	operationLogH *handler.OperationLogHandler,
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
		Notification: notificationH,
		OperationLog: operationLogH,
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
