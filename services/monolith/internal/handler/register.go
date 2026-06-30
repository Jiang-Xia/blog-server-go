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

	// rpg — Nest RpgController
	if deps.RPG != nil {
		rpgGroup := v1.Group("/rpg")
		rpgGroup.POST("/sign", jwtRequired, deps.RPG.SignIn)
		rpgGroup.GET("/sign-info", jwtRequired, deps.RPG.SignInfo)
		rpgGroup.GET("/status", jwtRequired, deps.RPG.Status)
		rpgGroup.GET("/hit-records", jwtRequired, deps.RPG.HitRecords)
		rpgGroup.GET("/level-rewards", deps.RPG.LevelRewards)
		rpgGroup.GET("/leaderboard", deps.RPG.Leaderboard)
		rpgGroup.GET("/ban-status", jwtRequired, deps.RPG.BanStatus)
		rpgGroup.GET("/my-achievements", jwtRequired, deps.RPG.MyAchievements)
		rpgGroup.GET("/quests", deps.RPG.Quests)
		rpgGroup.GET("/my-quests", jwtRequired, deps.RPG.MyQuests)
		rpgGroup.POST("/quest/claim", jwtRequired, deps.RPG.ClaimQuest)
		rpgGroup.GET("/my-buffs", jwtRequired, deps.RPG.MyBuffs)
		rpgGroup.POST("/buff/:id/activate", jwtRequired, deps.RPG.ActivateBuff)
		rpgGroup.POST("/buff/:id/deactivate", jwtRequired, deps.RPG.DeactivateBuff)
		rpgGroup.GET("/lottery/pool", deps.RPG.LotteryPool)
		rpgGroup.POST("/lottery/draw", jwtRequired, deps.RPG.LotteryDraw)
		rpgGroup.GET("/lottery/history", jwtRequired, deps.RPG.LotteryHistory)
		rpgGroup.GET("/lottery/tickets", jwtRequired, deps.RPG.LotteryTickets)
		rpgGroup.GET("/inventory", jwtRequired, deps.RPG.Inventory)
		rpgGroup.GET("/loadout", jwtRequired, deps.RPG.Loadout)
		rpgGroup.POST("/loadout/equip", jwtRequired, deps.RPG.EquipLoadout)
		rpgGroup.POST("/loadout/unequip", jwtRequired, deps.RPG.UnequipLoadout)
		rpgGroup.GET("/pets", jwtRequired, deps.RPG.Pets)
		rpgGroup.GET("/pets/catalog", deps.RPG.PetCatalog)
		rpgGroup.POST("/pets/summon", jwtRequired, deps.RPG.SummonPet)
		rpgGroup.POST("/pets/exchange", jwtRequired, deps.RPG.ExchangePet)
		rpgGroup.PATCH("/pets/:id/rename", jwtRequired, deps.RPG.RenamePet)
		rpgGroup.GET("/activities/current", deps.RPG.CurrentActivities)
		rpgGroup.POST("/activities/share-poster", jwtRequired, deps.RPG.SharePoster)
		rpgGroup.GET("/weather-buff", deps.RPG.WeatherBuff)
		rpgGroup.GET("/guilds", deps.RPG.ListGuilds)
		rpgGroup.GET("/guild/my", jwtRequired, deps.RPG.MyGuild)
		rpgGroup.GET("/guild/:id", deps.RPG.GuildDetail)
		rpgGroup.POST("/guild/create", jwtRequired, deps.RPG.CreateGuild)
		rpgGroup.POST("/guild/join", jwtRequired, deps.RPG.JoinGuild)
		rpgGroup.POST("/guild/leave", jwtRequired, deps.RPG.LeaveGuild)
		rpgGroup.POST("/article/tip", jwtRequired, deps.RPG.TipArticle)
		rpgGroup.POST("/social/cheer", jwtRequired, deps.RPG.Cheer)
		rpgGroup.POST("/social/egg", jwtRequired, deps.RPG.Egg)
		rpgGroup.POST("/social/flower", jwtRequired, deps.RPG.Flower)
		rpgGroup.POST("/recharge/create", jwtRequired, deps.RPG.CreateRecharge)
		rpgGroup.GET("/recharge/status", jwtRequired, deps.RPG.RechargeStatus)
	}

	// admin/rpg — Nest RpgAdminController（全路由需 JWT）
	if deps.RPGAdmin != nil {
		adminRpg := v1.Group("/admin/rpg", jwtRequired)
		adminRpg.GET("/achievements", deps.RPGAdmin.ListAchievements)
		adminRpg.POST("/achievements", deps.RPGAdmin.CreateAchievement)
		adminRpg.PATCH("/achievements/:id", deps.RPGAdmin.UpdateAchievement)
		adminRpg.DELETE("/achievements/:id", deps.RPGAdmin.DeleteAchievement)
		adminRpg.GET("/quests", deps.RPGAdmin.ListQuests)
		adminRpg.POST("/quests", deps.RPGAdmin.CreateQuest)
		adminRpg.PATCH("/quests/:id", deps.RPGAdmin.UpdateQuest)
		adminRpg.DELETE("/quests/:id", deps.RPGAdmin.DeleteQuest)
		adminRpg.GET("/lottery/pool", deps.RPGAdmin.ListLotteryPool)
		adminRpg.POST("/lottery/pool", deps.RPGAdmin.CreateLotteryPool)
		adminRpg.PATCH("/lottery/pool/:id", deps.RPGAdmin.UpdateLotteryPool)
		adminRpg.DELETE("/lottery/pool/:id", deps.RPGAdmin.DeleteLotteryPool)
		adminRpg.GET("/lottery/records", deps.RPGAdmin.ListLotteryRecords)
		adminRpg.GET("/users", deps.RPGAdmin.ListUsers)
		adminRpg.POST("/users/:uid/currency", deps.RPGAdmin.RechargeCurrency)
		adminRpg.POST("/users/:uid/currency/deduct", deps.RPGAdmin.DeductCurrency)
		adminRpg.POST("/users/:uid/unban", deps.RPGAdmin.UnbanUser)
		adminRpg.GET("/users/:uid", deps.RPGAdmin.GetUserDetail)
		adminRpg.GET("/stats", deps.RPGAdmin.Stats)
		adminRpg.GET("/items", deps.RPGAdmin.ListItems)
		adminRpg.POST("/items", deps.RPGAdmin.CreateItem)
		adminRpg.POST("/items/upload-asset", deps.RPGAdmin.UploadItemAsset)
		adminRpg.DELETE("/items/asset", deps.RPGAdmin.DeleteItemAsset)
		adminRpg.PATCH("/items/:id", deps.RPGAdmin.UpdateItem)
		adminRpg.DELETE("/items/:id", deps.RPGAdmin.DeleteItem)
		adminRpg.GET("/activities", deps.RPGAdmin.ListActivities)
		adminRpg.POST("/activities", deps.RPGAdmin.CreateActivity)
		adminRpg.PATCH("/activities/:id", deps.RPGAdmin.UpdateActivity)
		adminRpg.DELETE("/activities/:id", deps.RPGAdmin.DeleteActivity)
		adminRpg.GET("/guilds", deps.RPGAdmin.ListGuilds)
		adminRpg.DELETE("/guilds/:id", deps.RPGAdmin.DeleteGuild)
		adminRpg.GET("/guilds/:id/members", deps.RPGAdmin.ListGuildMembers)
		adminRpg.DELETE("/guilds/:id/members/:uid", deps.RPGAdmin.RemoveGuildMember)
		adminRpg.GET("/tips", deps.RPGAdmin.ListTips)
		adminRpg.GET("/social-logs", deps.RPGAdmin.ListSocialLogs)
	}

	// profile public — Nest ProfileController
	if deps.RPGProfile != nil {
		v1.GET("/user/public/:uid", deps.RPGProfile.PublicProfile)
		v1.GET("/user/public/:uid/articles", deps.RPGProfile.PublicArticles)
		v1.GET("/user/public/:uid/collects", deps.RPGProfile.PublicCollects)
		v1.GET("/user/public/:uid/likes", deps.RPGProfile.PublicLikes)
		v1.GET("/rpg/public/status/batch", deps.RPGProfile.PublicRpgStatusBatch)
		v1.GET("/rpg/public/:uid/status", deps.RPGProfile.PublicRpgStatus)
	}

	// pay — Nest PayController
	if deps.Pay != nil {
		payGroup := v1.Group("/pay")
		payGroup.POST("/trade/create", deps.Pay.TradeCreate)
		payGroup.GET("/trade/query", deps.Pay.TradeQuery)
		payGroup.POST("/trade/refund", deps.Pay.TradeRefund)
		payGroup.POST("/trade/close", deps.Pay.TradeClose)
		payGroup.POST("/openid", deps.Pay.GetOpenID)
		payGroup.POST("/h5-open-mini", deps.Pay.H5OpenMini)
		payGroup.POST("/notice", deps.Pay.Notice)
	}

	// pay/order — Nest PayOrderController（需 JWT）
	if deps.PayOrder != nil {
		payOrder := v1.Group("/pay/order", jwtRequired)
		payOrder.POST("/create", deps.PayOrder.Create)
		payOrder.GET("/list", deps.PayOrder.List)
		payOrder.POST("/refund", deps.PayOrder.Refund)
		payOrder.POST("/close", deps.PayOrder.Close)
		payOrder.GET("/query", deps.PayOrder.Query)
		payOrder.POST("/delete", deps.PayOrder.Delete)
		payOrder.POST("/mark-recharge-fulfilled", deps.PayOrder.MarkRechargeFulfilled)
	}
}
