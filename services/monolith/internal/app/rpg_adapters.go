// Package app RPG handler 适配层：将 rpg 子包服务方法映射为 handler 期望的接口签名。
package app

import (
	"context"
	"strconv"

	blogrepo "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/blog/repo"
	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/pkg/publicprofile"
	"github.com/Jiang-Xia/blog-server-go/pkg/redisutil"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent/sensitivewordhit"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/handler"
	payrepo "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/pay/repo"
	paysvc "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/pay/service"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg"
	rpgachievement "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/achievement"
	rpgactivity "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/activity"
	rpgadmin "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/admin"
	rpgbuff "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/buff"
	rpgevent "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/event"
	rpgguild "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/guild"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/inventory"
	rpgleaderboard "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/leaderboard"
	rpglottery "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/lottery"
	rpgpet "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/pet"
	rpgprofile "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/profile"
	rpgquest "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/quest"
	rpgrecharge "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/recharge"
	rpgsocial "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/social"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/auth"
	"go.uber.org/zap"
)

func provideRPGGameplay(mod *rpg.Module, client *ent.Client, articles *blogrepo.ArticleRepo) *handler.RPGGameplay {
	if mod == nil {
		return &handler.RPGGameplay{}
	}
	return &handler.RPGGameplay{
		Quest:       questAdapter{mod.Quest},
		Achievement: achievementAdapter{mod.Achievement},
		Buff:        buffAdapter{mod.Buff},
		Lottery:     lotteryAdapter{mod.Lottery},
		Inventory:   inventoryAdapter{mod.Inventory},
		Pet:         petAdapter{mod.Pet},
		Activity:    activityAdapter{mod.Activity},
		WeatherBuff: weatherAdapter{ws: mod.Weather},
		Leaderboard: leaderboardAdapter{mod.Leaderboard},
		Guild:       guildAdapter{mod.Guild},
		Tip:         tipAdapter{mod.Tip},
		Social:      socialAdapter{mod.Social},
		HitRecords:  hitRecordsAdapter{client: client},
		Recharge:    rechargeAdapter{mod.Recharge},
	}
}

func provideRPGAdminHandler(mod *rpg.Module, jwt *auth.JWTService) *handler.RPGAdminHandler {
	if mod == nil || mod.Admin == nil {
		return handler.NewRPGAdminHandler(nil, jwt)
	}
	return handler.NewRPGAdminHandler(adminAdapter{mod.Admin}, jwt)
}

func provideRPGProfileHandler(mod *rpg.Module, articles *blogrepo.ArticleRepo, publicProfile *publicprofile.Repo) *handler.RPGProfileHandler {
	if mod == nil || mod.Profile == nil {
		return handler.NewRPGProfileHandler(nil)
	}
	return handler.NewRPGProfileHandler(profileAdapter{svc: mod.Profile, articles: articles, publicProfile: publicProfile})
}

func providePayOrderServiceWithRecharge(
	orderRepo *payrepo.PayOrderRepo,
	pay *paysvc.PayService,
	mod *rpg.Module,
	log *zap.Logger,
) *paysvc.PayOrderService {
	var recharge paysvc.RechargeManualFulfiller
	if mod != nil && mod.Recharge != nil {
		recharge = mod.Recharge
		paysvc.RegisterPayPaidCallback(func(ctx context.Context, order *ent.PayOrder) error {
			_, err := mod.Recharge.TryFulfillOrderRecord(ctx, order)
			return err
		})
	}
	return paysvc.NewPayOrderService(orderRepo, pay, recharge, log)
}

func provideRPGEventHandlersFull(mod *rpg.Module, redis *redisutil.Store) rpgevent.Handlers {
	if mod == nil {
		return rpgevent.Handlers{Redis: redis}
	}
	return rpgevent.Handlers{
		Core:        mod.Rpg,
		Level:       mod.Level,
		Achievement: mod.Achievement,
		Quest:       mod.Quest,
		Reputation:  mod.Reputation,
		Redis:       redis,
	}
}

// --- adapters ---

type questAdapter struct{ *rpgquest.Service }

func (a questAdapter) GetActiveQuests(ctx context.Context, questType string) (interface{}, error) {
	return a.ListActiveQuests(ctx, questType)
}
func (a questAdapter) GetUserQuests(ctx context.Context, uid int) (interface{}, error) {
	return a.GetMyQuests(ctx, uid)
}
func (a questAdapter) ClaimReward(ctx context.Context, uid int, questCode string) (interface{}, error) {
	m, err := a.Service.ClaimReward(ctx, uid, questCode)
	return m, err
}

type achievementAdapter struct{ *rpgachievement.Service }

func (a achievementAdapter) GetUserAchievements(ctx context.Context, uid int) (interface{}, error) {
	return a.GetMyAchievements(ctx, uid)
}

type buffAdapter struct{ *rpgbuff.Service }

func (a buffAdapter) GetUserActiveBuffs(ctx context.Context, uid int) (interface{}, error) {
	return a.GetMyBuffs(ctx, uid)
}
func (a buffAdapter) ActivateBuff(ctx context.Context, uid, id int) (interface{}, error) {
	return nil, a.Service.ActivateBuff(ctx, uid, id)
}
func (a buffAdapter) DeactivateBuff(ctx context.Context, uid, id int) (interface{}, error) {
	return nil, a.Service.DeactivateBuff(ctx, uid, id)
}

type lotteryAdapter struct{ *rpglottery.Service }

func (a lotteryAdapter) GetPool(ctx context.Context) (interface{}, error) { return a.Service.GetPool(ctx) }
func (a lotteryAdapter) GetMyTickets(ctx context.Context, uid int) (int, error) {
	return a.Service.GetTickets(ctx, uid)
}
func (a lotteryAdapter) GetDrawHistory(ctx context.Context, uid, page, pageSize int) (interface{}, error) {
	limit := pageSize
	if limit <= 0 {
		limit = 20
	}
	return a.Service.GetHistory(ctx, uid, limit)
}
func (a lotteryAdapter) Draw(ctx context.Context, uid, count int, currency string) (interface{}, error) {
	useTicket := currency == "" || currency == "ticket"
	if count <= 0 {
		count = 1
	}
	results := make([]map[string]interface{}, 0, count)
	for i := 0; i < count; i++ {
		r, err := a.Service.Draw(ctx, uid, useTicket)
		if err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	if count == 1 {
		return results[0], nil
	}
	return map[string]interface{}{"results": results}, nil
}

type inventoryAdapter struct{ *inventory.Service }

func (a inventoryAdapter) GetUserInventory(ctx context.Context, uid int, itemTypes []string) (interface{}, error) {
	items, err := a.GetInventory(ctx, uid)
	if err != nil {
		return nil, err
	}
	if len(itemTypes) == 0 {
		return items, nil
	}
	allowed := map[string]struct{}{}
	for _, t := range itemTypes {
		allowed[t] = struct{}{}
	}
	filtered := make([]map[string]interface{}, 0)
	for _, it := range items {
		if t, ok := it["itemType"].(string); ok {
			if _, ok := allowed[t]; ok {
				filtered = append(filtered, it)
			}
		}
	}
	return filtered, nil
}
func (a inventoryAdapter) GetLoadoutDetail(ctx context.Context, uid int) (interface{}, error) {
	return a.Service.GetLoadoutDetail(ctx, uid)
}
func (a inventoryAdapter) EquipItem(ctx context.Context, uid int, slot string, body map[string]interface{}) (interface{}, error) {
	code, _ := body["itemCode"].(string)
	var petID *int
	if v, ok := body["petId"].(float64); ok {
		p := int(v)
		petID = &p
	}
	err := a.Equip(ctx, uid, inventory.LoadoutSlot(slot), code, petID)
	return map[string]bool{"success": err == nil}, err
}
func (a inventoryAdapter) UnequipItem(ctx context.Context, uid int, slot string) (interface{}, error) {
	err := a.Unequip(ctx, uid, inventory.LoadoutSlot(slot))
	return map[string]bool{"success": err == nil}, err
}

type petAdapter struct{ *rpgpet.Service }

func (a petAdapter) ListMyPets(ctx context.Context, uid int) (interface{}, error) { return a.ListPets(ctx, uid) }
func (a petAdapter) GetCatalog(ctx context.Context) (interface{}, error)         { return a.Service.GetCatalog(ctx) }
func (a petAdapter) SummonPet(ctx context.Context, uid int, itemCode string) (interface{}, error) {
	return a.Summon(ctx, uid, itemCode)
}
func (a petAdapter) ExchangePet(ctx context.Context, uid int, petCode string) (interface{}, error) {
	return a.Exchange(ctx, uid, petCode)
}
func (a petAdapter) RenamePet(ctx context.Context, uid, petID int, nickname string) (interface{}, error) {
	err := a.Rename(ctx, uid, petID, nickname)
	return map[string]bool{"success": err == nil}, err
}

type activityAdapter struct{ *rpgactivity.Service }

func (a activityAdapter) GetCurrentActivitiesOverview(ctx context.Context) (interface{}, error) {
	acts, err := a.GetCurrentActivities(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"activities": acts}, nil
}
func (a activityAdapter) RecordPosterShare(ctx context.Context, uid int, activityCode string) (interface{}, error) {
	a.SharePoster(ctx, uid, activityCode)
	return map[string]bool{"success": true}, nil
}

type weatherAdapter struct{ ws *rpgbuff.WeatherService }

func (a weatherAdapter) GetWeatherBuff(ctx context.Context, city string) (interface{}, error) {
	return a.ws.GetWeatherBuff(ctx, city)
}

type leaderboardAdapter struct{ *rpgleaderboard.Service }

func (a leaderboardAdapter) GetLeaderboard(ctx context.Context, scoreType, period string, limit int) (interface{}, error) {
	return a.Service.GetLeaderboard(ctx, rpgleaderboard.ScoreType(scoreType), rpgleaderboard.Period(period), limit)
}

type guildAdapter struct{ *rpgguild.Service }

func (a guildAdapter) ListGuilds(ctx context.Context, page, pageSize int, _ string) (interface{}, error) {
	return a.List(ctx, page, pageSize)
}
func (a guildAdapter) GetMyGuild(ctx context.Context, uid int) (interface{}, error) { return a.GetMy(ctx, uid) }
func (a guildAdapter) GetGuildDetail(ctx context.Context, id int) (interface{}, error) {
	return a.GetDetail(ctx, id, 0)
}
func (a guildAdapter) CreateGuild(ctx context.Context, uid int, name, announcement string) (interface{}, error) {
	return a.Create(ctx, uid, name, announcement)
}
func (a guildAdapter) JoinGuild(ctx context.Context, uid, guildID int) (interface{}, error) {
	return a.Join(ctx, uid, guildID)
}
func (a guildAdapter) LeaveGuild(ctx context.Context, uid int) (interface{}, error) {
	err := a.Leave(ctx, uid)
	return map[string]bool{"success": err == nil}, err
}

type tipAdapter struct{ *rpgsocial.TipService }

func (a tipAdapter) TipArticle(ctx context.Context, uid int, articleID string, amount int) (interface{}, error) {
	aid, err := strconv.Atoi(articleID)
	if err != nil {
		return nil, err
	}
	return a.TipService.TipArticle(ctx, uid, aid, amount)
}

type socialAdapter struct{ *rpgsocial.InteractService }

func (a socialAdapter) Interact(ctx context.Context, uid, targetUID int, action string) (interface{}, error) {
	switch action {
	case "cheer":
		return a.Cheer(ctx, uid, targetUID)
	case "egg":
		return a.Egg(ctx, uid, targetUID)
	case "flower":
		return a.Flower(ctx, uid, targetUID)
	default:
		return nil, errcode.WithMessage(errcode.InvalidParam, "未知互动类型")
	}
}

type hitRecordsAdapter struct{ client *ent.Client }

func (a hitRecordsAdapter) GetHitRecords(ctx context.Context, uid, page, pageSize int) (interface{}, error) {
	if a.client == nil {
		return map[string]interface{}{"list": []interface{}{}, "pagination": map[string]int{"total": 0, "page": 1, "pageSize": pageSize}}, nil
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize
	total, err := a.client.SensitiveWordHit.Query().Where(sensitivewordhit.UIDEQ(uid)).Count(ctx)
	if err != nil {
		return nil, err
	}
	list, err := a.client.SensitiveWordHit.Query().
		Where(sensitivewordhit.UIDEQ(uid)).
		Order(ent.Desc(sensitivewordhit.FieldCreateTime)).
		Offset(offset).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, err
	}
	totalPages := 0
	if pageSize > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}
	return map[string]interface{}{
		"list": list,
		"pagination": map[string]int{
			"total": total, "page": page, "pageSize": pageSize, "totalPages": totalPages,
		},
	}, nil
}

type rechargeAdapter struct{ *rpgrecharge.Service }

func (a rechargeAdapter) CreateRecharge(ctx context.Context, uid int, amountYuan float64) (interface{}, error) {
	return a.Service.CreateRecharge(ctx, uid, amountYuan)
}
func (a rechargeAdapter) GetRechargeStatus(ctx context.Context, uid int, outTradeNo string) (interface{}, error) {
	return a.Service.GetStatus(ctx, uid, outTradeNo)
}

type profileAdapter struct {
	svc           *rpgprofile.Service
	articles      *blogrepo.ArticleRepo
	publicProfile *publicprofile.Repo
}

func (a profileAdapter) GetPublicProfile(ctx context.Context, uid int) (interface{}, error) {
	return a.svc.GetPublicProfile(ctx, uid)
}
func (a profileAdapter) GetPublicRpgStatus(ctx context.Context, uid int) (interface{}, error) {
	return a.svc.GetPublicRpgStatus(ctx, uid)
}

func (a profileAdapter) ParsePublicRpgUIDs(raw string) []int {
	return rpgprofile.ParsePublicRpgUids(raw)
}
func (a profileAdapter) GetPublicRpgLevelsBatch(ctx context.Context, uids []int) (interface{}, error) {
	return a.svc.GetPublicRpgLevelsBatch(ctx, uids)
}
func (a profileAdapter) GetPublicArticles(ctx context.Context, uid, page, pageSize int) (interface{}, error) {
	if a.articles == nil {
		return map[string]interface{}{"list": []interface{}{}}, nil
	}
	list, err := a.articles.ListPublishedByAuthor(ctx, uid)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"list": list}, nil
}
func (a profileAdapter) GetPublicCollectArticles(ctx context.Context, uid, page, pageSize int) (interface{}, error) {
	if a.publicProfile == nil {
		return map[string]interface{}{"list": []interface{}{}, "pagination": map[string]int{"total": 0, "page": page, "pageSize": pageSize}}, nil
	}
	rows, total, err := a.publicProfile.ListCollectArticles(ctx, uid, page, pageSize)
	if err != nil {
		return nil, err
	}
	res := publicprofile.BuildListResult(rows, total, page, pageSize)
	return map[string]interface{}{"list": res.List, "pagination": res.Pagination}, nil
}
func (a profileAdapter) GetPublicLikeArticles(ctx context.Context, uid, page, pageSize int) (interface{}, error) {
	if a.publicProfile == nil {
		return map[string]interface{}{"list": []interface{}{}, "pagination": map[string]int{"total": 0, "page": page, "pageSize": pageSize}}, nil
	}
	rows, total, err := a.publicProfile.ListLikeArticles(ctx, uid, page, pageSize)
	if err != nil {
		return nil, err
	}
	res := publicprofile.BuildListResult(rows, total, page, pageSize)
	return map[string]interface{}{"list": res.List, "pagination": res.Pagination}, nil
}

type adminAdapter struct{ *rpgadmin.Service }

func pageFromQuery(q map[string]string) (page, pageSize int) {
	page, _ = strconv.Atoi(q["page"])
	pageSize, _ = strconv.Atoi(q["pageSize"])
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return page, pageSize
}

func (a adminAdapter) ListAchievements(ctx context.Context, query map[string]string) (interface{}, error) {
	p, ps := pageFromQuery(query)
	return a.Service.ListAchievements(ctx, p, ps)
}
func (a adminAdapter) CreateAchievement(ctx context.Context, body map[string]interface{}) (interface{}, error) {
	return a.Service.CreateAchievement(ctx, body)
}
func (a adminAdapter) UpdateAchievement(ctx context.Context, id string, body map[string]interface{}) (interface{}, error) {
	qid, _ := strconv.Atoi(id)
	return a.Service.UpdateAchievement(ctx, qid, body)
}
func (a adminAdapter) DeleteAchievement(ctx context.Context, id string) (interface{}, error) {
	qid, _ := strconv.Atoi(id)
	return a.Service.DeleteAchievement(ctx, qid)
}
func (a adminAdapter) ListQuests(ctx context.Context, query map[string]string) (interface{}, error) {
	p, ps := pageFromQuery(query)
	return a.Service.ListQuests(ctx, p, ps)
}
func (a adminAdapter) CreateQuest(ctx context.Context, body map[string]interface{}) (interface{}, error) {
	return a.Service.CreateQuestFromBody(ctx, body)
}
func (a adminAdapter) UpdateQuest(ctx context.Context, id string, body map[string]interface{}) (interface{}, error) {
	qid, _ := strconv.Atoi(id)
	return a.Service.UpdateQuestFromBody(ctx, qid, body)
}
func (a adminAdapter) DeleteQuest(ctx context.Context, id string) (interface{}, error) {
	qid, _ := strconv.Atoi(id)
	return map[string]bool{"success": true}, a.Service.DeleteQuest(ctx, qid)
}
func (a adminAdapter) ListLotteryPool(ctx context.Context, query map[string]string) (interface{}, error) {
	p, ps := pageFromQuery(query)
	return a.Service.ListLotteryPool(ctx, p, ps)
}
func (a adminAdapter) CreateLotteryPool(ctx context.Context, body map[string]interface{}) (interface{}, error) {
	return a.Service.CreateLotteryPoolFromBody(ctx, body)
}
func (a adminAdapter) UpdateLotteryPool(ctx context.Context, id string, body map[string]interface{}) (interface{}, error) {
	qid, _ := strconv.Atoi(id)
	return a.Service.UpdateLotteryPoolFromBody(ctx, qid, body)
}
func (a adminAdapter) DeleteLotteryPool(ctx context.Context, id string) (interface{}, error) {
	qid, _ := strconv.Atoi(id)
	return a.Service.DeleteLotteryPool(ctx, qid)
}
func (a adminAdapter) ListLotteryRecords(ctx context.Context, query map[string]string) (interface{}, error) {
	p, ps := pageFromQuery(query)
	return a.Service.ListLotteryRecords(ctx, p, ps)
}
func (a adminAdapter) ListUserRpgData(ctx context.Context, query map[string]string) (interface{}, error) {
	p, ps := pageFromQuery(query)
	return a.Service.ListUsers(ctx, p, ps)
}
func (a adminAdapter) GetUserRpgDetail(ctx context.Context, uid string) (interface{}, error) {
	id, _ := strconv.Atoi(uid)
	return a.Service.GetUserRpg(ctx, id)
}
func (a adminAdapter) RechargeCurrency(ctx context.Context, uid string, body map[string]interface{}, _ int) (interface{}, error) {
	id, _ := strconv.Atoi(uid)
	amount := int(body["amount"].(float64))
	bal, err := a.Service.RechargeDiamonds(ctx, id, amount)
	return map[string]int{"currency": bal}, err
}
func (a adminAdapter) DeductCurrency(ctx context.Context, uid string, body map[string]interface{}, _ int) (interface{}, error) {
	id, _ := strconv.Atoi(uid)
	amount := int(body["amount"].(float64))
	bal, err := a.Service.DeductDiamonds(ctx, id, amount)
	return map[string]int{"currency": bal}, err
}
func (a adminAdapter) UnbanUser(ctx context.Context, uid string, _ int) (interface{}, error) {
	id, _ := strconv.Atoi(uid)
	return a.Service.UnbanUser(ctx, id)
}
func (a adminAdapter) GetStats(ctx context.Context) (interface{}, error) { return a.Service.GetStats(ctx) }
func (a adminAdapter) ListItems(ctx context.Context, query map[string]string) (interface{}, error) {
	p, ps := pageFromQuery(query)
	return a.Service.ListItems(ctx, p, ps)
}
func (a adminAdapter) CreateItem(ctx context.Context, body map[string]interface{}) (interface{}, error) {
	return a.Service.CreateItemFromBody(ctx, body)
}
func (a adminAdapter) UpdateItem(ctx context.Context, id string, body map[string]interface{}) (interface{}, error) {
	iid, _ := strconv.Atoi(id)
	return a.Service.UpdateItemFromBody(ctx, iid, body)
}
func (a adminAdapter) DeleteItem(ctx context.Context, id string) (interface{}, error) {
	iid, _ := strconv.Atoi(id)
	return map[string]bool{"success": true}, a.Service.DeleteItem(ctx, iid)
}
func (a adminAdapter) UploadItemAsset(ctx context.Context, icon, assetType string, data []byte, filename string) (interface{}, error) {
	return a.Service.UploadItemAsset(ctx, icon, assetType, data, filename, "")
}
func (a adminAdapter) DeleteItemAsset(ctx context.Context, icon, assetType string) (interface{}, error) {
	return a.Service.DeleteItemAsset(ctx, icon, assetType)
}
func (a adminAdapter) ListActivities(ctx context.Context, query map[string]string) (interface{}, error) {
	p, ps := pageFromQuery(query)
	return a.Service.ListActivities(ctx, p, ps)
}
func (a adminAdapter) CreateActivity(ctx context.Context, body map[string]interface{}) (interface{}, error) {
	return a.Service.CreateActivityFromBody(ctx, body)
}
func (a adminAdapter) UpdateActivity(ctx context.Context, id string, body map[string]interface{}) (interface{}, error) {
	aid, _ := strconv.Atoi(id)
	return body, a.Service.UpdateActivity(ctx, aid, body)
}
func (a adminAdapter) DeleteActivity(ctx context.Context, id string) (interface{}, error) {
	aid, _ := strconv.Atoi(id)
	return map[string]bool{"success": true}, a.Service.DeleteActivity(ctx, aid)
}
func (a adminAdapter) ListGuilds(ctx context.Context, query map[string]string) (interface{}, error) {
	p, ps := pageFromQuery(query)
	return a.Service.ListGuilds(ctx, p, ps)
}
func (a adminAdapter) DeleteGuild(ctx context.Context, id string) (interface{}, error) {
	gid, _ := strconv.Atoi(id)
	return a.Service.DeleteGuild(ctx, gid)
}
func (a adminAdapter) ListGuildMembers(ctx context.Context, id string) (interface{}, error) {
	gid, _ := strconv.Atoi(id)
	return a.Service.ListGuildMembers(ctx, gid)
}
func (a adminAdapter) RemoveGuildMember(ctx context.Context, guildID, uid string) (interface{}, error) {
	gid, _ := strconv.Atoi(guildID)
	muid, _ := strconv.Atoi(uid)
	return a.Service.RemoveGuildMember(ctx, gid, muid)
}
func (a adminAdapter) ListTips(ctx context.Context, query map[string]string) (interface{}, error) {
	p, ps := pageFromQuery(query)
	return a.Service.ListTips(ctx, p, ps)
}
func (a adminAdapter) ListSocialLogs(ctx context.Context, query map[string]string) (interface{}, error) {
	p, ps := pageFromQuery(query)
	return a.Service.ListSocialLogs(ctx, p, ps)
}
