// Package handler 博客域路由（文章/互动/资源/通知/WS）。
//
// 鉴权约定：v1 组挂 Permission（RBAC + isPublic 白名单）；jwtRequired 为 Bearer JWT 强制登录并校验账号状态。
// WebSocket 在 API 前缀外注册 /realtime，鉴权见 WSHandler（query token 或 Authorization）。
package handler

import (
	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/middleware"
	"github.com/cloudwego/hertz/pkg/app/server"
)

// RegisterBlog 注册博客域 HTTP 与 WebSocket 路由。
func RegisterBlog(r *server.Hertz, cfg *config.Config, deps RegisterDeps) {
	// WS 不经 v1 Permission 组，升级时自行校验 JWT。
	if deps.WS != nil {
		deps.WS.Register(r)
	}

	perm := middleware.Permission(deps.Permission)
	jwtRequired := middleware.RequiredJWT(deps.JWT, deps.UserRepo)
	v1 := r.Group(cfg.App.APIPrefix, perm)

	// --- 站内通知（均需 JWT） ---
	notify := v1.Group("/notification")
	notify.GET("/list", jwtRequired, deps.Notification.List)
	notify.GET("/unread-count", jwtRequired, deps.Notification.UnreadCount)
	notify.GET("/since", jwtRequired, deps.Notification.Since)
	notify.PATCH("/read", jwtRequired, deps.Notification.MarkRead)

	// --- 开发调试（仅 development + JWT） ---
	if cfg.App.Env == "development" && deps.DevPush != nil {
		dev := v1.Group("/dev")
		dev.POST("/ws-push", jwtRequired, deps.DevPush.TestPush)
		dev.POST("/ws-push-redis", jwtRequired, deps.DevPush.TestPushRedis)
		dev.POST("/event-publish", jwtRequired, deps.DevPush.TestEvent)
	}

	// --- 文章：列表/详情公开；写操作需 JWT ---
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

	// --- 分类 / 标签：读公开，写需 JWT ---
	category := v1.Group("/category")
	category.POST("", jwtRequired, deps.Category.Create)
	category.GET("", deps.Category.List)
	category.GET("/:id", deps.Category.Get)
	category.PATCH("/:id", jwtRequired, deps.Category.Update)
	category.DELETE("/:id", jwtRequired, deps.Category.Delete)

	tag := v1.Group("/tag")
	tag.POST("", jwtRequired, deps.Tag.Create)
	tag.GET("", deps.Tag.List)
	tag.GET("/:id/article", deps.Tag.GetArticles)
	tag.GET("/:id", deps.Tag.Get)
	tag.PATCH("/:id", jwtRequired, deps.Tag.Update)
	tag.DELETE("/:id", jwtRequired, deps.Tag.Delete)

	// --- 评论 / 回复：创建删除需 JWT，列表多公开 ---
	comment := v1.Group("/comment")
	comment.POST("/create", jwtRequired, deps.Comment.Create)
	comment.DELETE("/delete", jwtRequired, deps.Comment.Delete)
	comment.GET("/findAll", deps.Comment.FindAll)
	comment.GET("/admin", deps.Comment.Admin)
	comment.GET("/my-list", jwtRequired, deps.Comment.MyList)
	comment.GET("/on-my-articles", jwtRequired, deps.Comment.OnMyArticles)

	reply := v1.Group("/reply")
	reply.POST("/create", jwtRequired, deps.Reply.Create)
	reply.DELETE("/delete", jwtRequired, deps.Reply.Delete)
	reply.GET("/findAll", deps.Reply.FindAll)
	reply.GET("/my-list", jwtRequired, deps.Reply.MyList)

	// --- 点赞 / 收藏：部分公开读，写操作需 JWT ---
	v1.POST("/like", deps.Like.Update)
	v1.GET("/like/check", jwtRequired, deps.Like.Check)
	v1.GET("/like/my-ids", jwtRequired, deps.Like.MyIDs)

	v1.POST("/collect", jwtRequired, deps.Collect.Toggle)
	v1.DELETE("/collect/:id", jwtRequired, deps.Collect.Delete)
	v1.GET("/collect/list", jwtRequired, deps.Collect.List)
	v1.GET("/collect/check", jwtRequired, deps.Collect.Check)
	v1.GET("/collect/count", deps.Collect.Count)

	// --- 留言板 / 友链：匿名可发，删除需 JWT ---
	v1.POST("/msgboard", deps.Msgboard.Create)
	v1.GET("/msgboard", deps.Msgboard.List)
	v1.POST("/msgboard/delete", jwtRequired, deps.Msgboard.Delete)

	v1.POST("/link", deps.Link.Create)
	v1.GET("/link", deps.Link.List)
	v1.GET("/link/:id", deps.Link.Get)
	v1.PATCH("/link/:id", deps.Link.Update)
	v1.DELETE("/link", jwtRequired, deps.Link.Delete)

	// --- 大文件分片与资源：上传需 JWT ---
	fileGroup := v1.Group("/file")
	fileGroup.POST("/uploadBigFile", jwtRequired, deps.File.UploadBigFile)
	fileGroup.POST("/uploadBigFile/merge", jwtRequired, deps.File.MergeBigFile)
	fileGroup.GET("/uploadBigFile/checkFile", jwtRequired, deps.File.CheckBigFile)

	res := v1.Group("/resources")
	res.GET("/daily-img", deps.Resources.DailyImg)
	res.GET("/weather", deps.Resources.Weather)
	res.GET("/baidutongji", jwtRequired, deps.Resources.BaiduTongJi)
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
