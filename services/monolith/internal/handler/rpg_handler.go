// Package handler C 端 RPG HTTP 端点，路径对齐 Nest RpgController。
package handler

import (
	"context"
	"strconv"

	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/pkg/response"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/auth"
	"github.com/cloudwego/hertz/pkg/app"
)

// RechargeCreator RPG 充值下单（由 rpg/recharge 实现）。
type RechargeCreator interface {
	CreateRecharge(ctx context.Context, uid int, amountYuan float64) (interface{}, error)
	GetRechargeStatus(ctx context.Context, uid int, outTradeNo string) (interface{}, error)
}

// RPGGameplay 扩展 RPG 玩法服务（quest/lottery 等，并行子模块注入）。
type RPGGameplay struct {
	Quest        interface{ GetActiveQuests(ctx context.Context, questType string) (interface{}, error); GetUserQuests(ctx context.Context, uid int) (interface{}, error); ClaimReward(ctx context.Context, uid int, questCode string) (interface{}, error) }
	Achievement  interface{ GetUserAchievements(ctx context.Context, uid int) (interface{}, error) }
	Buff         interface{ GetUserActiveBuffs(ctx context.Context, uid int) (interface{}, error); ActivateBuff(ctx context.Context, uid, id int) (interface{}, error); DeactivateBuff(ctx context.Context, uid, id int) (interface{}, error) }
	Lottery      interface{ GetPool(ctx context.Context) (interface{}, error); Draw(ctx context.Context, uid, count int, currency string) (interface{}, error); GetDrawHistory(ctx context.Context, uid, page, pageSize int) (interface{}, error); GetMyTickets(ctx context.Context, uid int) (int, error) }
	Inventory    interface{ GetUserInventory(ctx context.Context, uid int, itemTypes []string) (interface{}, error); GetLoadoutDetail(ctx context.Context, uid int) (interface{}, error); EquipItem(ctx context.Context, uid int, slot string, body map[string]interface{}) (interface{}, error); UnequipItem(ctx context.Context, uid int, slot string) (interface{}, error) }
	Pet          interface{ ListMyPets(ctx context.Context, uid int) (interface{}, error); GetCatalog(ctx context.Context) (interface{}, error); SummonPet(ctx context.Context, uid int, itemCode string) (interface{}, error); ExchangePet(ctx context.Context, uid int, petCode string) (interface{}, error); RenamePet(ctx context.Context, uid, petID int, nickname string) (interface{}, error) }
	Activity     interface{ GetCurrentActivitiesOverview(ctx context.Context) (interface{}, error); RecordPosterShare(ctx context.Context, uid int, activityCode string) (interface{}, error) }
	WeatherBuff  interface{ GetWeatherBuff(ctx context.Context, city string) (interface{}, error) }
	Leaderboard  interface{ GetLeaderboard(ctx context.Context, scoreType, period string, limit int) (interface{}, error) }
	Guild        interface{ ListGuilds(ctx context.Context, page, pageSize int, keyword string) (interface{}, error); GetMyGuild(ctx context.Context, uid int) (interface{}, error); GetGuildDetail(ctx context.Context, id int) (interface{}, error); CreateGuild(ctx context.Context, uid int, name, announcement string) (interface{}, error); JoinGuild(ctx context.Context, uid, guildID int) (interface{}, error); LeaveGuild(ctx context.Context, uid int) (interface{}, error) }
	Tip          interface{ TipArticle(ctx context.Context, uid int, articleID string, amount int) (interface{}, error) }
	Social       interface{ Interact(ctx context.Context, uid, targetUID int, action string) (interface{}, error) }
	HitRecords   interface{ GetHitRecords(ctx context.Context, uid, page, pageSize int) (interface{}, error) }
	Recharge     RechargeCreator
}

// RPGHandler /rpg/* 路由。
type RPGHandler struct {
	mod  *rpg.Module
	game *RPGGameplay
	jwt  *auth.JWTService
}

// NewRPGHandler 构造 RPGHandler。
func NewRPGHandler(mod *rpg.Module, game *RPGGameplay, jwt *auth.JWTService) *RPGHandler {
	if game == nil {
		game = &RPGGameplay{}
	}
	return &RPGHandler{mod: mod, game: game, jwt: jwt}
}

func (h *RPGHandler) uid(ctx context.Context, c *app.RequestContext) int {
	return articleUID(ctx, c, h.jwt)
}

func rpgNotReady(ctx context.Context, c *app.RequestContext) {
	response.Error(ctx, c, errcode.WithMessage(errcode.InternalError, "RPG 功能模块加载中"))
}

func (h *RPGHandler) SignIn(ctx context.Context, c *app.RequestContext) {
	uid := h.uid(ctx, c)
	if uid == 0 {
		response.Error(ctx, c, errcode.Unauthorized)
		return
	}
	if err := h.mod.Punishment.AssertNotBanned(ctx, uid); err != nil {
		response.FromError(ctx, c, err)
		return
	}
	data, err := h.mod.Sign.SignIn(ctx, uid)
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGHandler) SignInfo(ctx context.Context, c *app.RequestContext) {
	uid := h.uid(ctx, c)
	if uid == 0 {
		response.Error(ctx, c, errcode.Unauthorized)
		return
	}
	data, err := h.mod.Sign.GetSignInfo(ctx, uid)
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGHandler) Status(ctx context.Context, c *app.RequestContext) {
	uid := h.uid(ctx, c)
	if uid == 0 {
		response.Error(ctx, c, errcode.Unauthorized)
		return
	}
	data, err := h.mod.Rpg.GetFullStatus(ctx, uid)
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGHandler) HitRecords(ctx context.Context, c *app.RequestContext) {
	uid := h.uid(ctx, c)
	if uid == 0 {
		response.Error(ctx, c, errcode.Unauthorized)
		return
	}
	if h.game.HitRecords == nil {
		rpgNotReady(ctx, c)
		return
	}
	page := queryInt(c, "page", 1)
	pageSize := queryInt(c, "pageSize", 10)
	data, err := h.game.HitRecords.GetHitRecords(ctx, uid, page, pageSize)
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGHandler) LevelRewards(ctx context.Context, c *app.RequestContext) {
	data, err := h.mod.Level.GetLevelRewards(ctx, 10000)
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGHandler) Leaderboard(ctx context.Context, c *app.RequestContext) {
	if h.game.Leaderboard == nil {
		rpgNotReady(ctx, c)
		return
	}
	scoreType := string(c.Query("type"))
	if scoreType == "" {
		scoreType = "exp"
	}
	if scoreType == "signDays" {
		scoreType = "signDays"
	}
	period := string(c.Query("period"))
	if period == "" {
		period = "total"
	}
	limit := queryInt(c, "limit", 10)
	data, err := h.game.Leaderboard.GetLeaderboard(ctx, scoreType, period, limit)
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGHandler) BanStatus(ctx context.Context, c *app.RequestContext) {
	uid := h.uid(ctx, c)
	if uid == 0 {
		response.Error(ctx, c, errcode.Unauthorized)
		return
	}
	data, err := h.mod.Punishment.GetBanStatus(ctx, uid)
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGHandler) MyAchievements(ctx context.Context, c *app.RequestContext) {
	uid := h.uid(ctx, c)
	if uid == 0 {
		response.Error(ctx, c, errcode.Unauthorized)
		return
	}
	if h.game.Achievement == nil {
		rpgNotReady(ctx, c)
		return
	}
	data, err := h.game.Achievement.GetUserAchievements(ctx, uid)
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGHandler) Quests(ctx context.Context, c *app.RequestContext) {
	if h.game.Quest == nil {
		rpgNotReady(ctx, c)
		return
	}
	data, err := h.game.Quest.GetActiveQuests(ctx, "daily")
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGHandler) MyQuests(ctx context.Context, c *app.RequestContext) {
	uid := h.uid(ctx, c)
	if uid == 0 {
		response.Error(ctx, c, errcode.Unauthorized)
		return
	}
	if h.game.Quest == nil {
		rpgNotReady(ctx, c)
		return
	}
	data, err := h.game.Quest.GetUserQuests(ctx, uid)
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGHandler) ClaimQuest(ctx context.Context, c *app.RequestContext) {
	uid := h.uid(ctx, c)
	if uid == 0 {
		response.Error(ctx, c, errcode.Unauthorized)
		return
	}
	if h.game.Quest == nil {
		rpgNotReady(ctx, c)
		return
	}
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	data, err := h.game.Quest.ClaimReward(ctx, uid, strField(body, "questCode"))
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGHandler) MyBuffs(ctx context.Context, c *app.RequestContext) {
	uid := h.uid(ctx, c)
	if uid == 0 {
		response.Error(ctx, c, errcode.Unauthorized)
		return
	}
	if h.game.Buff == nil {
		rpgNotReady(ctx, c)
		return
	}
	data, err := h.game.Buff.GetUserActiveBuffs(ctx, uid)
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGHandler) ActivateBuff(ctx context.Context, c *app.RequestContext) {
	uid := h.uid(ctx, c)
	if uid == 0 {
		response.Error(ctx, c, errcode.Unauthorized)
		return
	}
	if h.game.Buff == nil {
		rpgNotReady(ctx, c)
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	data, err := h.game.Buff.ActivateBuff(ctx, uid, id)
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGHandler) DeactivateBuff(ctx context.Context, c *app.RequestContext) {
	uid := h.uid(ctx, c)
	if uid == 0 {
		response.Error(ctx, c, errcode.Unauthorized)
		return
	}
	if h.game.Buff == nil {
		rpgNotReady(ctx, c)
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	data, err := h.game.Buff.DeactivateBuff(ctx, uid, id)
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGHandler) LotteryPool(ctx context.Context, c *app.RequestContext) {
	if h.game.Lottery == nil {
		rpgNotReady(ctx, c)
		return
	}
	data, err := h.game.Lottery.GetPool(ctx)
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGHandler) LotteryDraw(ctx context.Context, c *app.RequestContext) {
	uid := h.uid(ctx, c)
	if uid == 0 {
		response.Error(ctx, c, errcode.Unauthorized)
		return
	}
	if h.game.Lottery == nil {
		rpgNotReady(ctx, c)
		return
	}
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	count := intField(body, "count")
	if count <= 0 {
		count = 1
	}
	currency := strField(body, "currency")
	if currency == "" {
		currency = "ticket"
	}
	data, err := h.game.Lottery.Draw(ctx, uid, count, currency)
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGHandler) LotteryHistory(ctx context.Context, c *app.RequestContext) {
	uid := h.uid(ctx, c)
	if uid == 0 {
		response.Error(ctx, c, errcode.Unauthorized)
		return
	}
	if h.game.Lottery == nil {
		rpgNotReady(ctx, c)
		return
	}
	page := queryInt(c, "page", 1)
	pageSize := queryInt(c, "pageSize", 20)
	data, err := h.game.Lottery.GetDrawHistory(ctx, uid, page, pageSize)
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGHandler) LotteryTickets(ctx context.Context, c *app.RequestContext) {
	uid := h.uid(ctx, c)
	if uid == 0 {
		response.Error(ctx, c, errcode.Unauthorized)
		return
	}
	if h.game.Lottery == nil {
		rpgNotReady(ctx, c)
		return
	}
	tickets, err := h.game.Lottery.GetMyTickets(ctx, uid)
	if err != nil {
		handleAdminResult(ctx, c, nil, err)
		return
	}
	handleAdminResult(ctx, c, map[string]int{"tickets": tickets}, nil)
}

func (h *RPGHandler) Inventory(ctx context.Context, c *app.RequestContext) {
	uid := h.uid(ctx, c)
	if uid == 0 {
		response.Error(ctx, c, errcode.Unauthorized)
		return
	}
	if h.game.Inventory == nil {
		rpgNotReady(ctx, c)
		return
	}
	itemType := string(c.Query("itemType"))
	var types []string
	if itemType != "" {
		types = []string{itemType}
	}
	items, err := h.game.Inventory.GetUserInventory(ctx, uid, types)
	if err != nil {
		handleAdminResult(ctx, c, nil, err)
		return
	}
	handleAdminResult(ctx, c, map[string]interface{}{"items": items}, nil)
}

func (h *RPGHandler) Loadout(ctx context.Context, c *app.RequestContext) {
	uid := h.uid(ctx, c)
	if uid == 0 {
		response.Error(ctx, c, errcode.Unauthorized)
		return
	}
	if h.game.Inventory == nil {
		rpgNotReady(ctx, c)
		return
	}
	data, err := h.game.Inventory.GetLoadoutDetail(ctx, uid)
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGHandler) EquipLoadout(ctx context.Context, c *app.RequestContext) {
	uid := h.uid(ctx, c)
	if uid == 0 {
		response.Error(ctx, c, errcode.Unauthorized)
		return
	}
	if h.game.Inventory == nil {
		rpgNotReady(ctx, c)
		return
	}
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	data, err := h.game.Inventory.EquipItem(ctx, uid, strField(body, "slot"), body)
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGHandler) UnequipLoadout(ctx context.Context, c *app.RequestContext) {
	uid := h.uid(ctx, c)
	if uid == 0 {
		response.Error(ctx, c, errcode.Unauthorized)
		return
	}
	if h.game.Inventory == nil {
		rpgNotReady(ctx, c)
		return
	}
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	data, err := h.game.Inventory.UnequipItem(ctx, uid, strField(body, "slot"))
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGHandler) Pets(ctx context.Context, c *app.RequestContext) {
	uid := h.uid(ctx, c)
	if uid == 0 {
		response.Error(ctx, c, errcode.Unauthorized)
		return
	}
	if h.game.Pet == nil {
		rpgNotReady(ctx, c)
		return
	}
	data, err := h.game.Pet.ListMyPets(ctx, uid)
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGHandler) PetCatalog(ctx context.Context, c *app.RequestContext) {
	if h.game.Pet == nil {
		rpgNotReady(ctx, c)
		return
	}
	data, err := h.game.Pet.GetCatalog(ctx)
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGHandler) SummonPet(ctx context.Context, c *app.RequestContext) {
	uid := h.uid(ctx, c)
	if uid == 0 {
		response.Error(ctx, c, errcode.Unauthorized)
		return
	}
	if h.game.Pet == nil {
		rpgNotReady(ctx, c)
		return
	}
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	data, err := h.game.Pet.SummonPet(ctx, uid, strField(body, "itemCode"))
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGHandler) ExchangePet(ctx context.Context, c *app.RequestContext) {
	uid := h.uid(ctx, c)
	if uid == 0 {
		response.Error(ctx, c, errcode.Unauthorized)
		return
	}
	if h.game.Pet == nil {
		rpgNotReady(ctx, c)
		return
	}
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	data, err := h.game.Pet.ExchangePet(ctx, uid, strField(body, "petCode"))
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGHandler) RenamePet(ctx context.Context, c *app.RequestContext) {
	uid := h.uid(ctx, c)
	if uid == 0 {
		response.Error(ctx, c, errcode.Unauthorized)
		return
	}
	if h.game.Pet == nil {
		rpgNotReady(ctx, c)
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	data, err := h.game.Pet.RenamePet(ctx, uid, id, strField(body, "nickname"))
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGHandler) CurrentActivities(ctx context.Context, c *app.RequestContext) {
	if h.game.Activity == nil {
		rpgNotReady(ctx, c)
		return
	}
	data, err := h.game.Activity.GetCurrentActivitiesOverview(ctx)
	if err != nil {
		handleAdminResult(ctx, c, nil, err)
		return
	}
	handleAdminResult(ctx, c, data, nil)
}

func (h *RPGHandler) SharePoster(ctx context.Context, c *app.RequestContext) {
	uid := h.uid(ctx, c)
	if uid == 0 {
		response.Error(ctx, c, errcode.Unauthorized)
		return
	}
	if h.game.Activity == nil {
		rpgNotReady(ctx, c)
		return
	}
	var body map[string]interface{}
	_ = c.Bind(&body)
	data, err := h.game.Activity.RecordPosterShare(ctx, uid, strField(body, "activityCode"))
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGHandler) WeatherBuff(ctx context.Context, c *app.RequestContext) {
	if h.game.WeatherBuff == nil {
		rpgNotReady(ctx, c)
		return
	}
	city := string(c.Query("city"))
	if city == "" {
		city = "北京"
	}
	data, err := h.game.WeatherBuff.GetWeatherBuff(ctx, city)
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGHandler) ListGuilds(ctx context.Context, c *app.RequestContext) {
	if h.game.Guild == nil {
		rpgNotReady(ctx, c)
		return
	}
	page := queryInt(c, "page", 1)
	pageSize := queryInt(c, "pageSize", 20)
	data, err := h.game.Guild.ListGuilds(ctx, page, pageSize, string(c.Query("keyword")))
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGHandler) MyGuild(ctx context.Context, c *app.RequestContext) {
	uid := h.uid(ctx, c)
	if uid == 0 {
		response.Error(ctx, c, errcode.Unauthorized)
		return
	}
	if h.game.Guild == nil {
		rpgNotReady(ctx, c)
		return
	}
	data, err := h.game.Guild.GetMyGuild(ctx, uid)
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGHandler) GuildDetail(ctx context.Context, c *app.RequestContext) {
	if h.game.Guild == nil {
		rpgNotReady(ctx, c)
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	data, err := h.game.Guild.GetGuildDetail(ctx, id)
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGHandler) CreateGuild(ctx context.Context, c *app.RequestContext) {
	uid := h.uid(ctx, c)
	if uid == 0 {
		response.Error(ctx, c, errcode.Unauthorized)
		return
	}
	if h.game.Guild == nil {
		rpgNotReady(ctx, c)
		return
	}
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	data, err := h.game.Guild.CreateGuild(ctx, uid, strField(body, "name"), strField(body, "announcement"))
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGHandler) JoinGuild(ctx context.Context, c *app.RequestContext) {
	uid := h.uid(ctx, c)
	if uid == 0 {
		response.Error(ctx, c, errcode.Unauthorized)
		return
	}
	if h.game.Guild == nil {
		rpgNotReady(ctx, c)
		return
	}
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	guildID := intField(body, "guildId")
	data, err := h.game.Guild.JoinGuild(ctx, uid, guildID)
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGHandler) LeaveGuild(ctx context.Context, c *app.RequestContext) {
	uid := h.uid(ctx, c)
	if uid == 0 {
		response.Error(ctx, c, errcode.Unauthorized)
		return
	}
	if h.game.Guild == nil {
		rpgNotReady(ctx, c)
		return
	}
	data, err := h.game.Guild.LeaveGuild(ctx, uid)
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGHandler) TipArticle(ctx context.Context, c *app.RequestContext) {
	uid := h.uid(ctx, c)
	if uid == 0 {
		response.Error(ctx, c, errcode.Unauthorized)
		return
	}
	if h.game.Tip == nil {
		rpgNotReady(ctx, c)
		return
	}
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	data, err := h.game.Tip.TipArticle(ctx, uid, strField(body, "articleId"), intField(body, "amount"))
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGHandler) Cheer(ctx context.Context, c *app.RequestContext) {
	h.socialInteract(ctx, c, "cheer")
}

func (h *RPGHandler) Egg(ctx context.Context, c *app.RequestContext) {
	h.socialInteract(ctx, c, "egg")
}

func (h *RPGHandler) Flower(ctx context.Context, c *app.RequestContext) {
	h.socialInteract(ctx, c, "flower")
}

func (h *RPGHandler) socialInteract(ctx context.Context, c *app.RequestContext, action string) {
	uid := h.uid(ctx, c)
	if uid == 0 {
		response.Error(ctx, c, errcode.Unauthorized)
		return
	}
	if h.game.Social == nil {
		rpgNotReady(ctx, c)
		return
	}
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	targetUID := intField(body, "targetUid")
	if targetUID <= 0 {
		response.Error(ctx, c, errcode.WithMessage(errcode.InvalidParam, "目标用户无效"))
		return
	}
	data, err := h.game.Social.Interact(ctx, uid, targetUID, action)
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGHandler) CreateRecharge(ctx context.Context, c *app.RequestContext) {
	uid := h.uid(ctx, c)
	if uid == 0 {
		response.Error(ctx, c, errcode.Unauthorized)
		return
	}
	if err := h.mod.Punishment.AssertNotBanned(ctx, uid); err != nil {
		response.FromError(ctx, c, err)
		return
	}
	if h.game.Recharge == nil {
		rpgNotReady(ctx, c)
		return
	}
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	amount, ok := toFloat(body["amountYuan"])
	if !ok {
		response.Error(ctx, c, errcode.WithMessage(errcode.InvalidParam, "amountYuan 无效"))
		return
	}
	data, err := h.game.Recharge.CreateRecharge(ctx, uid, amount)
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGHandler) RechargeStatus(ctx context.Context, c *app.RequestContext) {
	uid := h.uid(ctx, c)
	if uid == 0 {
		response.Error(ctx, c, errcode.Unauthorized)
		return
	}
	outTradeNo := string(c.Query("out_trade_no"))
	if outTradeNo == "" {
		response.Error(ctx, c, errcode.WithMessage(errcode.InvalidParam, "商户订单号不能为空"))
		return
	}
	if h.game.Recharge == nil {
		rpgNotReady(ctx, c)
		return
	}
	data, err := h.game.Recharge.GetRechargeStatus(ctx, uid, outTradeNo)
	handleAdminResult(ctx, c, data, err)
}

func toFloat(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case string:
		f, err := strconv.ParseFloat(n, 64)
		return f, err == nil
	default:
		return 0, false
	}
}
