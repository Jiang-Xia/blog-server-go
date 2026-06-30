// Package handler 注册全部 HTTP 路由（health / user / captcha / pub）。
package handler

import (
	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/middleware"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/auth"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/repo"
	"github.com/cloudwego/hertz/pkg/app/server"
)

// RegisterDeps 路由注册依赖，由 wire 装配后传入。
type RegisterDeps struct {
	Health     *HealthHandler
	User       *UserHandler
	Captcha    *CaptchaHandler
	Pub        *PubHandler
	JWT        *auth.JWTService
	UserRepo   *repo.UserRepo
	Permission middleware.PermissionDeps
}

// RegisterAll 注册 health、user、captcha、pub 路由；v1 组挂载 Permission 中间件。
func RegisterAll(r *server.Hertz, cfg *config.Config, deps RegisterDeps) {
	r.GET("/health", deps.Health.OK)
	r.GET(cfg.App.APIPrefix+"/health", deps.Health.OK)

	perm := middleware.Permission(deps.Permission)
	jwtRequired := middleware.RequiredJWT(deps.JWT, deps.UserRepo)

	v1 := r.Group(cfg.App.APIPrefix, perm)

	// captcha
	v1.GET("/captcha", deps.Captcha.Get)
	v1.POST("/captcha/verify", deps.Captcha.Verify)

	// pub
	v1.GET("/pub/stats", deps.Pub.Stats)

	// user — 路径与 Nest @Controller('user') 一致
	user := v1.Group("/user")
	user.GET("/authCode", deps.User.AuthCode)
	user.POST("/register", deps.User.Register)
	user.POST("/login", deps.User.Login)
	user.GET("/refresh", deps.User.Refresh)
	user.GET("/info", jwtRequired, deps.User.Info)
	user.POST("/list", deps.User.List)
	user.PATCH("/status", jwtRequired, deps.User.UpdateStatus)
	user.PATCH("/edit", jwtRequired, deps.User.Edit)
	user.PATCH("/password", jwtRequired, deps.User.Password)
	user.POST("/resetPassword", deps.User.ResetPassword)
	user.DELETE("", jwtRequired, deps.User.Delete)
	user.POST("/email/sendCode", deps.User.SendEmailCode)
	user.POST("/email/register", deps.User.EmailRegister)
	user.POST("/email/login", deps.User.EmailLogin)
	user.GET("/auth/github", deps.User.GithubAuth)
	user.GET("/auth/github/callback", deps.User.GithubCallback)
	user.POST("/auth/ticket/exchange", deps.User.ExchangeOAuthTicket)
	user.POST("/auth/wechat/miniprogram", deps.User.WechatMiniProgramLogin)
	user.POST("/admin/create", jwtRequired, deps.User.AdminCreate)
	user.POST("/admin/update/:id", jwtRequired, deps.User.AdminUpdate)
}
