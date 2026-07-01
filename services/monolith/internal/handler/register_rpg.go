// Package handler RPG/支付域路由。
//
// 鉴权：C 端公开读（奖池/排行榜等）无 jwtRequired；/rpg 写与个人数据需 JWT；
// /admin/rpg 与 /pay/order 需 JWT；/pay/notice 为支付宝验签异步回调（无 JWT，靠签名校验）。
package handler

import (
	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/middleware"
	"github.com/cloudwego/hertz/pkg/app/server"
)

// RegisterRPG 注册 RPG、支付、公开主页路由。
func RegisterRPG(r *server.Hertz, cfg *config.Config, deps RegisterDeps) {
	perm := middleware.Permission(deps.Permission)
	jwtRequired := middleware.RequiredJWT(deps.JWT, deps.UserRepo)
	v1 := r.Group(cfg.App.APIPrefix, perm)

	// --- C 端 RPG：公开读 + JWT 写 ---
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

	// --- 管理端 RPG 配置（均需 JWT + RBAC） ---
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

	// --- 公开主页（无需 JWT） ---
	if deps.RPGProfile != nil {
		v1.GET("/user/public/:uid", deps.RPGProfile.PublicProfile)
		v1.GET("/user/public/:uid/articles", deps.RPGProfile.PublicArticles)
		v1.GET("/user/public/:uid/collects", deps.RPGProfile.PublicCollects)
		v1.GET("/user/public/:uid/likes", deps.RPGProfile.PublicLikes)
		v1.GET("/rpg/public/status/batch", deps.RPGProfile.PublicRpgStatusBatch)
		v1.GET("/rpg/public/:uid/status", deps.RPGProfile.PublicRpgStatus)
	}

	// --- 支付：C 端下单公开；notice 为支付宝回调（验签，无 JWT） ---
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

	// --- 管理端支付订单（均需 JWT） ---
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
