// Package handler 用户域路由注册依赖。
package handler

import (
	"github.com/Jiang-Xia/blog-server-go/services/user/internal/middleware"
	"github.com/Jiang-Xia/blog-server-go/services/user/internal/user/auth"
	"github.com/Jiang-Xia/blog-server-go/services/user/internal/user/repo"
)

// RegisterDeps 路由注册依赖，由 wire 装配后传入。
type RegisterDeps struct {
	Health       *HealthHandler
	User         *UserHandler
	Admin        *AdminHandler
	Captcha      *CaptchaHandler
	Sensitive    *SensitiveWordHandler
	OperationLog *OperationLogHandler
	JWT          *auth.JWTService
	UserRepo     *repo.UserRepo
	Permission   middleware.PermissionDeps
	OpLog        middleware.OperationLogDeps
}
