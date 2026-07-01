//go:generate go run github.com/google/wire/cmd/wire
//go:build wireinject
// +build wireinject

// Package app wire 依赖注入声明（make wire-user 生成 wire_gen.go）。
package app

import (
	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/logger"
	"github.com/Jiang-Xia/blog-server-go/services/user/internal/data"
	"github.com/Jiang-Xia/blog-server-go/services/user/internal/handler"
	"github.com/Jiang-Xia/blog-server-go/services/user/internal/operationlog"
	"github.com/Jiang-Xia/blog-server-go/services/user/internal/server"
	"github.com/Jiang-Xia/blog-server-go/services/user/internal/user/admin"
	"github.com/Jiang-Xia/blog-server-go/services/user/internal/user/auth"
	"github.com/Jiang-Xia/blog-server-go/services/user/internal/user/captcha"
	"github.com/Jiang-Xia/blog-server-go/services/user/internal/user/email"
	"github.com/Jiang-Xia/blog-server-go/services/user/internal/user/profile"
	"github.com/Jiang-Xia/blog-server-go/services/user/internal/user/repo"
	"github.com/Jiang-Xia/blog-server-go/services/user/internal/user/sensitive"
	"github.com/google/wire"
)

// InitializeApp 装配 user-service 全部依赖。
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
		captcha.NewService,
		handler.NewUserAppAdapter,
		provideUserHandler,
		handler.NewAdminHandler,
		provideCaptchaHandler,
		handler.NewSensitiveWordHandler,
		operationlog.NewService,
		handler.NewOperationLogHandler,
		handler.NewHealthHandler,
		provideRegisterDeps,
		provideUserGRPCServer,
		server.NewHTTPServer,
		NewApp,
	)
	return nil, nil
}
