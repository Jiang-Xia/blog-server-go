// Package handler ??????user / admin / captcha / sensitive-word / operation-log??
package handler

import (
	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/services/user/internal/middleware"
	"github.com/cloudwego/hertz/pkg/app/server"
)

// Register ?? health ???????user-service ??????
func Register(r *server.Hertz, cfg *config.Config, deps RegisterDeps) {
	r.GET("/health", deps.Health.OK)
	r.GET(cfg.App.APIPrefix+"/health", deps.Health.OK)
	registerUserRoutes(r, cfg, deps)
}

// RegisterUser ?????????monolith shim ???health ? monolith ????
func RegisterUser(r *server.Hertz, cfg *config.Config, deps RegisterDeps) {
	registerUserRoutes(r, cfg, deps)
}

func registerUserRoutes(r *server.Hertz, cfg *config.Config, deps RegisterDeps) {
	perm := middleware.Permission(deps.Permission)
	jwtRequired := middleware.RequiredJWT(deps.JWT, deps.UserRepo)
	v1 := r.Group(cfg.App.APIPrefix, perm)

	v1.GET("/captcha", deps.Captcha.Get)
	v1.POST("/captcha/verify", deps.Captcha.Verify)

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

	role := v1.Group("/role")
	role.GET("/menu-privilege-tree", deps.Admin.RoleMenuPrivilegeTree)
	role.POST("", deps.Admin.RoleCreate)
	role.GET("", deps.Admin.RoleList)
	role.GET("/:id/data-scope", deps.Admin.RoleGetDataScope)
	role.PUT("/:id/data-scope", deps.Admin.RoleUpdateDataScope)
	role.GET("/:id", deps.Admin.RoleGet)
	role.PATCH("/:id", deps.Admin.RoleUpdate)
	role.DELETE("/:id", deps.Admin.RoleDelete)

	dept := v1.Group("/dept")
	dept.POST("", deps.Admin.DeptCreate)
	dept.GET("/tree", deps.Admin.DeptTree)
	dept.GET("", deps.Admin.DeptList)
	dept.GET("/:id", deps.Admin.DeptGet)
	dept.PATCH("/:id", deps.Admin.DeptUpdate)
	dept.DELETE("/:id", deps.Admin.DeptDelete)

	priv := v1.Group("/privilege")
	priv.POST("", deps.Admin.PrivilegeCreate)
	priv.GET("", deps.Admin.PrivilegeList)
	priv.GET("/:id", deps.Admin.PrivilegeGet)
	priv.PATCH("/:id", deps.Admin.PrivilegeUpdate)
	priv.DELETE("/:id", deps.Admin.PrivilegeDelete)

	adminGroup := v1.Group("/admin")
	adminGroup.GET("/menu", deps.Admin.MenuList)
	adminGroup.POST("/menu", deps.Admin.MenuCreate)
	adminGroup.PATCH("/menu", deps.Admin.MenuUpdate)
	adminGroup.GET("/menu/detail", deps.Admin.MenuDetail)
	adminGroup.DELETE("/menu", jwtRequired, deps.Admin.MenuDelete)

	sw := v1.Group("/sensitive-word")
	sw.GET("", deps.Sensitive.List)
	sw.POST("", deps.Sensitive.Create)
	sw.POST("/batch", deps.Sensitive.BatchCreate)
	sw.GET("/hit", deps.Sensitive.ListHits)
	sw.POST("/hit/:id/approve", deps.Sensitive.Approve)
	sw.POST("/hit/:id/reject", deps.Sensitive.Reject)
	sw.PATCH("/:id", deps.Sensitive.Update)
	sw.DELETE("/:id", deps.Sensitive.Delete)

	opLog := v1.Group("/operation-log")
	opLog.GET("", deps.OperationLog.List)
	opLog.DELETE("/clean", deps.OperationLog.Clean)
}
