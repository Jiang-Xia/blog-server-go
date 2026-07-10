// Package handler 注册全部 HTTP 路由（health / user / blog / rpg / pub）。
//
// 按域拆分：RegisterUser（认证与 RBAC）、RegisterBlog（内容与互动）、RegisterRPG（玩法与支付）。
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
	Health       *HealthHandler
	User         *UserHandler
	Admin        *AdminHandler
	Captcha      *CaptchaHandler
	Pub          *PubHandler
	Sensitive    *SensitiveWordHandler
	Article      *ArticleHandler
	Category     *CategoryHandler
	Tag          *TagHandler
	Comment      *CommentHandler
	Reply        *ReplyHandler
	Like         *LikeHandler
	Collect      *CollectHandler
	Msgboard     *MsgboardHandler
	Link         *LinkHandler
	File         *FileHandler
	Resources    *ResourcesHandler
	Notification *NotificationHandler
	ScheduledTask *ScheduledTaskHandler
	Rag          *RagHandler
	OperationLog *OperationLogHandler
	WS           *WSHandler
	DevPush      *DevPushHandler
	RPG          *RPGHandler
	RPGAdmin     *RPGAdminHandler
	RPGProfile   *RPGProfileHandler
	Pay          *PayHandler
	PayOrder     *PayOrderHandler
	JWT          *auth.JWTService
	UserRepo     *repo.UserRepo
	Permission   middleware.PermissionDeps
	OpLog        middleware.OperationLogDeps
}

// RegisterAll 注册全部域路由（单体模式）。
func RegisterAll(r *server.Hertz, cfg *config.Config, deps RegisterDeps) {
	r.GET("/health", deps.Health.OK)
	r.GET(cfg.App.APIPrefix+"/health", deps.Health.OK)
	RegisterUser(r, cfg, deps)
	RegisterBlog(r, cfg, deps)
	RegisterRPG(r, cfg, deps)

	// 公开统计（isPublic，无 JWT）
	perm := middleware.Permission(deps.Permission)
	v1 := r.Group(cfg.App.APIPrefix, perm)
	v1.GET("/pub/stats", deps.Pub.Stats)
}
