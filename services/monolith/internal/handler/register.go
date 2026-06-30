// Package handler 注册全部 HTTP 路由（health / user / admin / captcha / pub）。
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
	OperationLog *OperationLogHandler
	WS           *WSHandler
	DevPush      *DevPushHandler
	JWT          *auth.JWTService
	UserRepo     *repo.UserRepo
	Permission   middleware.PermissionDeps
	OpLog        middleware.OperationLogDeps
}

// RegisterAll 注册 health、user、captcha、pub 路由；v1 组挂载 Permission 中间件。
func RegisterAll(r *server.Hertz, cfg *config.Config, deps RegisterDeps) {
	r.GET("/health", deps.Health.OK)
	r.GET(cfg.App.APIPrefix+"/health", deps.Health.OK)

	if deps.WS != nil {
		deps.WS.Register(r)
	}

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

	// role — Nest RoleController
	role := v1.Group("/role")
	role.GET("/menu-privilege-tree", deps.Admin.RoleMenuPrivilegeTree)
	role.POST("", deps.Admin.RoleCreate)
	role.GET("", deps.Admin.RoleList)
	role.GET("/:id/data-scope", deps.Admin.RoleGetDataScope)
	role.PUT("/:id/data-scope", deps.Admin.RoleUpdateDataScope)
	role.GET("/:id", deps.Admin.RoleGet)
	role.PATCH("/:id", deps.Admin.RoleUpdate)
	role.DELETE("/:id", deps.Admin.RoleDelete)

	// dept — Nest DeptController
	dept := v1.Group("/dept")
	dept.POST("", deps.Admin.DeptCreate)
	dept.GET("/tree", deps.Admin.DeptTree)
	dept.GET("", deps.Admin.DeptList)
	dept.GET("/:id", deps.Admin.DeptGet)
	dept.PATCH("/:id", deps.Admin.DeptUpdate)
	dept.DELETE("/:id", deps.Admin.DeptDelete)

	// privilege — Nest PrivilegeController
	priv := v1.Group("/privilege")
	priv.POST("", deps.Admin.PrivilegeCreate)
	priv.GET("", deps.Admin.PrivilegeList)
	priv.GET("/:id", deps.Admin.PrivilegeGet)
	priv.PATCH("/:id", deps.Admin.PrivilegeUpdate)
	priv.DELETE("/:id", deps.Admin.PrivilegeDelete)

	// admin/menu — Nest MenuController
	adminGroup := v1.Group("/admin")
	adminGroup.GET("/menu", deps.Admin.MenuList)
	adminGroup.POST("/menu", deps.Admin.MenuCreate)
	adminGroup.PATCH("/menu", deps.Admin.MenuUpdate)
	adminGroup.GET("/menu/detail", deps.Admin.MenuDetail)
	adminGroup.DELETE("/menu", jwtRequired, deps.Admin.MenuDelete)

	// sensitive-word — Nest SensitiveWordController
	sw := v1.Group("/sensitive-word")
	sw.GET("", deps.Sensitive.List)
	sw.POST("", deps.Sensitive.Create)
	sw.POST("/batch", deps.Sensitive.BatchCreate)
	sw.GET("/hit", deps.Sensitive.ListHits)
	sw.POST("/hit/:id/approve", deps.Sensitive.Approve)
	sw.POST("/hit/:id/reject", deps.Sensitive.Reject)
	sw.PATCH("/:id", deps.Sensitive.Update)
	sw.DELETE("/:id", deps.Sensitive.Delete)

	// notification — Nest NotificationController + since 骨架
	notify := v1.Group("/notification")
	notify.GET("/list", jwtRequired, deps.Notification.List)
	notify.GET("/unread-count", jwtRequired, deps.Notification.UnreadCount)
	notify.GET("/since", jwtRequired, deps.Notification.Since)
	notify.PATCH("/read", jwtRequired, deps.Notification.MarkRead)

	// dev — WS 冒烟（development）
	if cfg.App.Env == "development" && deps.DevPush != nil {
		dev := v1.Group("/dev")
		dev.POST("/ws-push", jwtRequired, deps.DevPush.TestPush)
		dev.POST("/ws-push-redis", jwtRequired, deps.DevPush.TestPushRedis)
		dev.POST("/event-publish", jwtRequired, deps.DevPush.TestEvent)
	}

	// operation-log — Nest OperationLogController
	opLog := v1.Group("/operation-log")
	opLog.GET("", deps.OperationLog.List)
	opLog.DELETE("/clean", deps.OperationLog.Clean)

	// article — Nest ArticleController
	article := v1.Group("/article")
	article.POST("/list", deps.Article.List)
	article.GET("/info", deps.Article.Info)
	article.POST("/create", jwtRequired, deps.Article.Create)
	article.POST("/edit", jwtRequired, deps.Article.Edit)
	article.DELETE("/delete", jwtRequired, deps.Article.Delete)
	article.POST("/views", deps.Article.Views)
	article.POST("/likes", deps.Article.Likes)
	article.PATCH("/disabled", deps.Article.Disabled)
	article.PATCH("/topping", deps.Article.Topping)
	article.GET("/my-list", jwtRequired, deps.Article.MyList)
	article.GET("/archives", deps.Article.Archives)
	article.GET("/related", deps.Article.Related)
	article.GET("/author-stats", jwtRequired, deps.Article.AuthorStats)
	article.GET("/statistics", deps.Article.Statistics)

	// category — Nest CategoryController
	category := v1.Group("/category")
	category.POST("", jwtRequired, deps.Category.Create)
	category.GET("", deps.Category.List)
	category.GET("/:id", deps.Category.Get)
	category.PATCH("/:id", jwtRequired, deps.Category.Update)
	category.DELETE("/:id", jwtRequired, deps.Category.Delete)

	// tag — Nest TagController
	tag := v1.Group("/tag")
	tag.POST("", jwtRequired, deps.Tag.Create)
	tag.GET("", deps.Tag.List)
	tag.GET("/:id/article", deps.Tag.GetArticles)
	tag.GET("/:id", deps.Tag.Get)
	tag.PATCH("/:id", jwtRequired, deps.Tag.Update)
	tag.DELETE("/:id", jwtRequired, deps.Tag.Delete)

	// comment — Nest CommentController
	comment := v1.Group("/comment")
	comment.POST("/create", jwtRequired, deps.Comment.Create)
	comment.DELETE("/delete", jwtRequired, deps.Comment.Delete)
	comment.GET("/findAll", deps.Comment.FindAll)
	comment.GET("/admin", deps.Comment.Admin)
	comment.GET("/my-list", jwtRequired, deps.Comment.MyList)
	comment.GET("/on-my-articles", jwtRequired, deps.Comment.OnMyArticles)

	// reply — Nest ReplyController
	reply := v1.Group("/reply")
	reply.POST("/create", jwtRequired, deps.Reply.Create)
	reply.DELETE("/delete", jwtRequired, deps.Reply.Delete)
	reply.GET("/findAll", deps.Reply.FindAll)
	reply.GET("/my-list", jwtRequired, deps.Reply.MyList)

	// like — Nest LikeController
	v1.POST("/like", deps.Like.Update)
	v1.GET("/like/check", jwtRequired, deps.Like.Check)
	v1.GET("/like/my-ids", jwtRequired, deps.Like.MyIDs)

	// collect — Nest CollectController
	v1.POST("/collect", jwtRequired, deps.Collect.Toggle)
	v1.DELETE("/collect/:id", jwtRequired, deps.Collect.Delete)
	v1.GET("/collect/list", jwtRequired, deps.Collect.List)
	v1.GET("/collect/check", jwtRequired, deps.Collect.Check)
	v1.GET("/collect/count", deps.Collect.Count)

	// msgboard — Nest MsgboardController
	v1.POST("/msgboard", deps.Msgboard.Create)
	v1.GET("/msgboard", deps.Msgboard.List)
	v1.POST("/msgboard/delete", jwtRequired, deps.Msgboard.Delete)

	// link — Nest LinkController
	v1.POST("/link", deps.Link.Create)
	v1.GET("/link", deps.Link.List)
	v1.GET("/link/:id", deps.Link.Get)
	v1.PATCH("/link/:id", deps.Link.Update)
	v1.DELETE("/link", jwtRequired, deps.Link.Delete)

	// file — Nest FileController（大文件分片）
	fileGroup := v1.Group("/file")
	fileGroup.POST("/uploadBigFile", jwtRequired, deps.File.UploadBigFile)
	fileGroup.POST("/uploadBigFile/merge", jwtRequired, deps.File.MergeBigFile)
	fileGroup.GET("/uploadBigFile/checkFile", jwtRequired, deps.File.CheckBigFile)

	// resources — Nest ResourcesController
	res := v1.Group("/resources")
	res.GET("/daily-img", deps.Resources.DailyImg)
	res.GET("/weather", deps.Resources.Weather)
	res.POST("/uploadFile", jwtRequired, deps.Resources.UploadFile)
	res.POST("/upload-media", jwtRequired, deps.Resources.UploadMedia)
	res.POST("/upload-media/register-avatar", deps.Resources.RegisterAvatar)
	res.GET("/files", deps.Resources.Files)
	res.GET("/register-avatars", deps.Resources.RegisterAvatars)
	res.GET("/file/:id", deps.Resources.GetFile)
	res.DELETE("/file", jwtRequired, deps.Resources.DeleteFile)
	res.POST("/folder", deps.Resources.CreateFolder)
	res.PATCH("/file", deps.Resources.UpdateFile)
}
