// Package app wire provider 集合。
package app

import (
	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/redisutil"
	"github.com/Jiang-Xia/blog-server-go/services/user/internal/handler"
	"github.com/Jiang-Xia/blog-server-go/services/user/internal/middleware"
	"github.com/Jiang-Xia/blog-server-go/services/user/internal/operationlog"
	"github.com/Jiang-Xia/blog-server-go/services/user/internal/user/auth"
	"github.com/Jiang-Xia/blog-server-go/services/user/internal/user/captcha"
	"github.com/Jiang-Xia/blog-server-go/services/user/internal/user/profile"
	"github.com/Jiang-Xia/blog-server-go/services/user/internal/user/repo"
	usersgrpc "github.com/Jiang-Xia/blog-server-go/services/user/internal/user/grpcserver"
	"github.com/redis/rueidis"
	"go.uber.org/zap"
)

func provideUserGRPCServer(profileSvc *profile.Service, jwt *auth.JWTService) *usersgrpc.Server {
	return usersgrpc.New(profileSvc, jwt)
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

func provideRegisterDeps(
	health *handler.HealthHandler,
	userH *handler.UserHandler,
	adminH *handler.AdminHandler,
	captchaH *handler.CaptchaHandler,
	sensitiveH *handler.SensitiveWordHandler,
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
		Sensitive:    sensitiveH,
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
