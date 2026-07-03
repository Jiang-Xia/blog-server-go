// Package repo RPG 域 Ent 仓储封装。
package repo

import (
	"context"
	"time"

	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent/rpg"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent/rpgactivity"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent/rpgarticletip"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent/rpgguild"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent/rpgitemconfig"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent/rpglevelreward"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent/rpglotterypool"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent/rpgquest"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent/rpguserachievement"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent/rpguserbuff"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent/rpguserguildmember"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent/rpguserinventory"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent/rpguserloadout"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent/rpguserlotteryrecord"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent/rpguserpet"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent/rpguserquestprogress"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent/rpgleaderboardsnapshot"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent/payorder"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent/rpgusersociallog"
)

// RpgRepo RPG 相关表读写。
type RpgRepo struct {
	client *ent.Client
}

// NewRpgRepo 构造 RpgRepo。
func NewRpgRepo(client *ent.Client) *RpgRepo {
	return &RpgRepo{client: client}
}

// WithTx 在 Ent 事务中执行 fn。
func (r *RpgRepo) WithTx(ctx context.Context, fn func(tx *ent.Tx) error) error {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if v := recover(); v != nil {
			_ = tx.Rollback()
			panic(v)
		}
	}()
	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return err
		}
		return err
	}
	return tx.Commit()
}

// --- Rpg ---

// FindRpgByUID 按 uid 查询 RPG 记录。
func (r *RpgRepo) FindRpgByUID(ctx context.Context, uid int) (*ent.Rpg, error) {
	return r.client.Rpg.Query().Where(rpg.UIDEQ(uid), rpg.IsDelete(false)).First(ctx)
}

// CreateRpg 创建 RPG 记录。
func (r *RpgRepo) CreateRpg(ctx context.Context, uid int) (*ent.Rpg, error) {
	return r.client.Rpg.Create().SetUID(uid).Save(ctx)
}

// UpdateRpg 更新 RPG 记录。
func (r *RpgRepo) UpdateRpg(ctx context.Context, row *ent.Rpg) (*ent.Rpg, error) {
	up := r.client.Rpg.UpdateOneID(row.ID).
		SetExp(row.Exp).
		SetLevel(row.Level).
		SetLifeValue(row.LifeValue).
		SetTotalSignDays(row.TotalSignDays).
		SetConsecutiveSignDays(row.ConsecutiveSignDays).
		SetSensitiveHitsCount(row.SensitiveHitsCount).
		SetZeroLifeCount(row.ZeroLifeCount).
		SetLotteryTickets(row.LotteryTickets).
		SetReputation(row.Reputation).
		SetLotteryPityCounter(row.LotteryPityCounter).
		SetLotteryLegendaryPityCounter(row.LotteryLegendaryPityCounter)
	if row.LastSignDate != nil {
		up.SetLastSignDate(*row.LastSignDate)
	} else {
		up.ClearLastSignDate()
	}
	if row.BanStartTime != nil {
		up.SetBanStartTime(*row.BanStartTime)
	} else {
		up.ClearBanStartTime()
	}
	if row.BanEndTime != nil {
		up.SetBanEndTime(*row.BanEndTime)
	} else {
		up.ClearBanEndTime()
	}
	if row.EffectJson != nil {
		up.SetEffectJson(*row.EffectJson)
	} else {
		up.ClearEffectJson()
	}
	return up.Save(ctx)
}

// ListRpgOrderBy 总榜排序查询。
func (r *RpgRepo) ListRpgOrderBy(ctx context.Context, field string, limit int) ([]*ent.Rpg, error) {
	q := r.client.Rpg.Query().Where(rpg.IsDelete(false))
	switch field {
	case "level":
		q = q.Order(ent.Desc(rpg.FieldLevel), ent.Asc(rpg.FieldCreateTime))
	case "reputation":
		q = q.Order(ent.Desc(rpg.FieldReputation), ent.Asc(rpg.FieldCreateTime))
	case "signDays":
		q = q.Order(ent.Desc(rpg.FieldTotalSignDays), ent.Asc(rpg.FieldCreateTime))
	default:
		q = q.Order(ent.Desc(rpg.FieldExp), ent.Asc(rpg.FieldCreateTime))
	}
	return q.Limit(limit).All(ctx)
}

// ListRpgByUIDs 批量查 RPG。
func (r *RpgRepo) ListRpgByUIDs(ctx context.Context, uids []int) ([]*ent.Rpg, error) {
	if len(uids) == 0 {
		return nil, nil
	}
	return r.client.Rpg.Query().Where(rpg.UIDIn(uids...), rpg.IsDelete(false)).All(ctx)
}

// --- Inventory ---

// ListInventoryByUID 用户背包列表。
func (r *RpgRepo) ListInventoryByUID(ctx context.Context, uid int) ([]*ent.RpgUserInventory, error) {
	return r.client.RpgUserInventory.Query().
		Where(rpguserinventory.UIDEQ(uid), rpguserinventory.QuantityGT(0)).
		All(ctx)
}

// FindInventoryByUIDAndCode 查背包单行。
func (r *RpgRepo) FindInventoryByUIDAndCode(ctx context.Context, uid int, code string) (*ent.RpgUserInventory, error) {
	return r.client.RpgUserInventory.Query().
		Where(rpguserinventory.UIDEQ(uid), rpguserinventory.ItemCodeEQ(code)).
		First(ctx)
}

// FindInventoryByUIDAndItemCode 查背包单行（FindInventoryByUIDAndCode 别名）。
func (r *RpgRepo) FindInventoryByUIDAndItemCode(ctx context.Context, uid int, itemCode string) (*ent.RpgUserInventory, error) {
	return r.FindInventoryByUIDAndCode(ctx, uid, itemCode)
}

// CreateInventory 创建背包行。
func (r *RpgRepo) CreateInventory(ctx context.Context, row *ent.RpgUserInventory) (*ent.RpgUserInventory, error) {
	b := r.client.RpgUserInventory.Create().
		SetUID(row.UID).
		SetItemCode(row.ItemCode).
		SetQuantity(row.Quantity).
		SetSource(row.Source).
		SetAcquiredAt(time.Now())
	if row.EffectJson != nil {
		b.SetEffectJson(*row.EffectJson)
	}
	return b.Save(ctx)
}

// UpdateInventoryQuantity 更新数量。
func (r *RpgRepo) UpdateInventoryQuantity(ctx context.Context, id, qty int) error {
	_, err := r.client.RpgUserInventory.UpdateOneID(id).SetQuantity(qty).Save(ctx)
	return err
}

// ListCurrencyLeaderboard 货币总榜。
func (r *RpgRepo) ListCurrencyLeaderboard(ctx context.Context, currencyCode string, limit int) ([]struct {
	UID      int
	Currency int
}, error) {
	rows, err := r.client.RpgUserInventory.Query().
		Where(rpguserinventory.ItemCodeEQ(currencyCode), rpguserinventory.QuantityGT(0)).
		Order(ent.Desc(rpguserinventory.FieldQuantity)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]struct {
		UID      int
		Currency int
	}, 0, len(rows))
	for _, row := range rows {
		out = append(out, struct {
			UID      int
			Currency int
		}{UID: row.UID, Currency: row.Quantity})
	}
	return out, nil
}

// --- Loadout ---

// FindLoadoutByUID 查装扮槽。
func (r *RpgRepo) FindLoadoutByUID(ctx context.Context, uid int) (*ent.RpgUserLoadout, error) {
	return r.client.RpgUserLoadout.Query().Where(rpguserloadout.UIDEQ(uid)).First(ctx)
}

// CreateLoadout 创建空装扮槽。
func (r *RpgRepo) CreateLoadout(ctx context.Context, uid int) (*ent.RpgUserLoadout, error) {
	return r.client.RpgUserLoadout.Create().SetUID(uid).Save(ctx)
}

// GetOrCreateLoadout 获取或创建装扮槽。
func (r *RpgRepo) GetOrCreateLoadout(ctx context.Context, uid int) (*ent.RpgUserLoadout, error) {
	row, err := r.FindLoadoutByUID(ctx, uid)
	if err == nil {
		return row, nil
	}
	if ent.IsNotFound(err) {
		return r.CreateLoadout(ctx, uid)
	}
	return nil, err
}

// UpdateLoadout 更新装扮槽。
func (r *RpgRepo) UpdateLoadout(ctx context.Context, row *ent.RpgUserLoadout) (*ent.RpgUserLoadout, error) {
	up := r.client.RpgUserLoadout.UpdateOneID(row.ID).SetUID(row.UID)
	if row.TitleCode != nil {
		up.SetTitleCode(*row.TitleCode)
	} else {
		up.ClearTitleCode()
	}
	if row.AvatarFrameCode != nil {
		up.SetAvatarFrameCode(*row.AvatarFrameCode)
	} else {
		up.ClearAvatarFrameCode()
	}
	if row.PetId != nil {
		up.SetPetId(*row.PetId)
	} else {
		up.ClearPetId()
	}
	return up.Save(ctx)
}

// --- ItemConfig ---

// FindItemConfigByCode 按 code 查物品配置。
func (r *RpgRepo) FindItemConfigByCode(ctx context.Context, code string) (*ent.RpgItemConfig, error) {
	return r.client.RpgItemConfig.Query().Where(rpgitemconfig.CodeEQ(code)).First(ctx)
}

// FindItemConfigByID 按 ID 查物品配置。
func (r *RpgRepo) FindItemConfigByID(ctx context.Context, id int) (*ent.RpgItemConfig, error) {
	return r.client.RpgItemConfig.Get(ctx, id)
}

// ListItemConfigsByCodes 批量查配置。
func (r *RpgRepo) ListItemConfigsByCodes(ctx context.Context, codes []string) ([]*ent.RpgItemConfig, error) {
	if len(codes) == 0 {
		return nil, nil
	}
	return r.client.RpgItemConfig.Query().Where(rpgitemconfig.CodeIn(codes...)).All(ctx)
}

// ListItemConfigsByType 按类型查配置。
func (r *RpgRepo) ListItemConfigsByType(ctx context.Context, itemType string, activeOnly bool) ([]*ent.RpgItemConfig, error) {
	q := r.client.RpgItemConfig.Query().Where(rpgitemconfig.ItemTypeEQ(itemType))
	if activeOnly {
		q = q.Where(rpgitemconfig.ActiveEQ(1))
	}
	return q.Order(ent.Asc(rpgitemconfig.FieldSort)).All(ctx)
}

// CreateItemConfig 创建物品配置。
func (r *RpgRepo) CreateItemConfig(ctx context.Context, row *ent.RpgItemConfig) (*ent.RpgItemConfig, error) {
	b := r.client.RpgItemConfig.Create().
		SetCode(row.Code).
		SetName(row.Name).
		SetItemType(row.ItemType).
		SetDescription(row.Description).
		SetCategory(row.Category).
		SetIcon(row.Icon).
		SetRarity(row.Rarity).
		SetSort(row.Sort).
		SetActive(row.Active).
		SetIsHidden(row.IsHidden)
	if row.EffectJson != nil {
		b.SetEffectJson(*row.EffectJson)
	}
	return b.Save(ctx)
}

// UpdateItemConfig 更新物品配置字段。
func (r *RpgRepo) UpdateItemConfig(ctx context.Context, id int, patch map[string]interface{}) error {
	up := r.client.RpgItemConfig.UpdateOneID(id)
	for k, v := range patch {
		switch k {
		case rpgitemconfig.FieldName:
			up.SetName(v.(string))
		case rpgitemconfig.FieldDescription:
			up.SetDescription(v.(string))
		case rpgitemconfig.FieldItemType:
			up.SetItemType(v.(string))
		case rpgitemconfig.FieldCategory:
			up.SetCategory(v.(string))
		case rpgitemconfig.FieldIcon:
			up.SetIcon(v.(string))
		case rpgitemconfig.FieldRarity:
			up.SetRarity(v.(string))
		case rpgitemconfig.FieldSort:
			up.SetSort(v.(int))
		case rpgitemconfig.FieldActive:
			up.SetActive(v.(int))
		case rpgitemconfig.FieldEffectJson:
			up.SetEffectJson(v.(string))
		}
	}
	_, err := up.Save(ctx)
	return err
}

// ListItemConfigsAdmin 管理端分页列表。
func (r *RpgRepo) ListItemConfigsAdmin(ctx context.Context, offset, limit int) ([]*ent.RpgItemConfig, int, error) {
	q := r.client.RpgItemConfig.Query()
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	rows, err := q.Order(ent.Asc(rpgitemconfig.FieldSort)).Offset(offset).Limit(limit).All(ctx)
	return rows, total, err
}

// DeleteItemConfig 删除物品配置。
func (r *RpgRepo) DeleteItemConfig(ctx context.Context, id int) error {
	return r.client.RpgItemConfig.DeleteOneID(id).Exec(ctx)
}

// --- Quest ---

// FindQuestByCode 按 code 查任务。
func (r *RpgRepo) FindQuestByCode(ctx context.Context, code string) (*ent.RpgQuest, error) {
	return r.client.RpgQuest.Query().Where(rpgquest.CodeEQ(code)).First(ctx)
}

// ListActiveQuests 活跃任务列表。
func (r *RpgRepo) ListActiveQuests(ctx context.Context, questType string) ([]*ent.RpgQuest, error) {
	q := r.client.RpgQuest.Query().Where(rpgquest.ActiveEQ(1))
	if questType != "" {
		q = q.Where(rpgquest.TypeEQ(questType))
	}
	return q.Order(ent.Asc(rpgquest.FieldSort)).All(ctx)
}

// CreateQuest 创建任务。
func (r *RpgRepo) CreateQuest(ctx context.Context, row *ent.RpgQuest) (*ent.RpgQuest, error) {
	b := r.client.RpgQuest.Create().
		SetCode(row.Code).
		SetName(row.Name).
		SetDescription(row.Description).
		SetType(row.Type).
		SetQuestSubtype(row.QuestSubtype).
		SetTargetAction(row.TargetAction).
		SetTargetCount(row.TargetCount).
		SetExpReward(row.ExpReward).
		SetHpReward(row.HpReward).
		SetCurrencyReward(row.CurrencyReward).
		SetSort(row.Sort).
		SetActive(row.Active)
	if row.EffectJson != nil {
		b.SetEffectJson(*row.EffectJson)
	}
	return b.Save(ctx)
}

// UpdateQuest 更新任务。
func (r *RpgRepo) UpdateQuest(ctx context.Context, id int, patch map[string]interface{}) error {
	up := r.client.RpgQuest.UpdateOneID(id)
	for k, v := range patch {
		switch k {
		case rpgquest.FieldName:
			up.SetName(v.(string))
		case rpgquest.FieldDescription:
			up.SetDescription(v.(string))
		case rpgquest.FieldTargetCount:
			up.SetTargetCount(v.(int))
		case rpgquest.FieldExpReward:
			up.SetExpReward(v.(int))
		case rpgquest.FieldHpReward:
			up.SetHpReward(v.(int))
		case rpgquest.FieldCurrencyReward:
			up.SetCurrencyReward(v.(int))
		case rpgquest.FieldSort:
			up.SetSort(v.(int))
		case rpgquest.FieldActive:
			up.SetActive(v.(int))
		case rpgquest.FieldEffectJson:
			up.SetEffectJson(v.(string))
		}
	}
	_, err := up.Save(ctx)
	return err
}

// ListQuestsAdmin 管理端任务列表。
func (r *RpgRepo) ListQuestsAdmin(ctx context.Context, offset, limit int) ([]*ent.RpgQuest, int, error) {
	q := r.client.RpgQuest.Query()
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	rows, err := q.Order(ent.Asc(rpgquest.FieldSort)).Offset(offset).Limit(limit).All(ctx)
	return rows, total, err
}

// DeleteQuest 删除任务。
func (r *RpgRepo) DeleteQuest(ctx context.Context, id int) error {
	return r.client.RpgQuest.DeleteOneID(id).Exec(ctx)
}

// --- Quest Progress ---

// FindQuestProgress 查单条进度。
func (r *RpgRepo) FindQuestProgress(ctx context.Context, uid int, questCode string, questDate time.Time) (*ent.RpgUserQuestProgress, error) {
	return r.client.RpgUserQuestProgress.Query().
		Where(
			rpguserquestprogress.UIDEQ(uid),
			rpguserquestprogress.QuestCodeEQ(questCode),
			rpguserquestprogress.QuestDateEQ(questDate),
		).
		First(ctx)
}

// ListQuestProgressByUIDAndDate 按 uid+日期查进度。
func (r *RpgRepo) ListQuestProgressByUIDAndDate(ctx context.Context, uid int, questDate time.Time) ([]*ent.RpgUserQuestProgress, error) {
	return r.client.RpgUserQuestProgress.Query().
		Where(rpguserquestprogress.UIDEQ(uid), rpguserquestprogress.QuestDateEQ(questDate)).
		All(ctx)
}

// SaveQuestProgress 创建或更新进度。
func (r *RpgRepo) SaveQuestProgress(ctx context.Context, row *ent.RpgUserQuestProgress) (*ent.RpgUserQuestProgress, error) {
	if row.ID > 0 {
		return r.client.RpgUserQuestProgress.UpdateOneID(row.ID).
			SetProgress(row.Progress).
			SetCompleted(row.Completed).
			SetClaimed(row.Claimed).
			Save(ctx)
	}
	return r.client.RpgUserQuestProgress.Create().
		SetUID(row.UID).
		SetQuestCode(row.QuestCode).
		SetProgress(row.Progress).
		SetCompleted(row.Completed).
		SetClaimed(row.Claimed).
		SetQuestDate(row.QuestDate).
		Save(ctx)
}

// --- Lottery ---

// FindLotteryPoolByItemCode 按 itemCode 查奖池行。
func (r *RpgRepo) FindLotteryPoolByItemCode(ctx context.Context, itemCode string) (*ent.RpgLotteryPool, error) {
	return r.client.RpgLotteryPool.Query().Where(rpglotterypool.ItemCodeEQ(itemCode)).First(ctx)
}

// ListActiveLotteryPool 活跃奖池。
func (r *RpgRepo) ListActiveLotteryPool(ctx context.Context) ([]*ent.RpgLotteryPool, error) {
	return r.client.RpgLotteryPool.Query().
		Where(rpglotterypool.ActiveEQ(1)).
		Order(ent.Asc(rpglotterypool.FieldSort)).
		All(ctx)
}

// CreateLotteryPool 创建奖池行。
func (r *RpgRepo) CreateLotteryPool(ctx context.Context, row *ent.RpgLotteryPool) (*ent.RpgLotteryPool, error) {
	return r.client.RpgLotteryPool.Create().
		SetItemCode(row.ItemCode).
		SetProbability(row.Probability).
		SetRarity(row.Rarity).
		SetSort(row.Sort).
		SetActive(row.Active).
		Save(ctx)
}

// UpdateLotteryPool 更新奖池。
func (r *RpgRepo) UpdateLotteryPool(ctx context.Context, id int, patch map[string]interface{}) error {
	up := r.client.RpgLotteryPool.UpdateOneID(id)
	for k, v := range patch {
		switch k {
		case rpglotterypool.FieldProbability:
			up.SetProbability(v.(float64))
		case rpglotterypool.FieldRarity:
			up.SetRarity(v.(string))
		case rpglotterypool.FieldSort:
			up.SetSort(v.(int))
		case rpglotterypool.FieldActive:
			up.SetActive(v.(int))
		}
	}
	_, err := up.Save(ctx)
	return err
}

// ListLotteryPoolAdmin 管理端奖池列表。
func (r *RpgRepo) ListLotteryPoolAdmin(ctx context.Context, offset, limit int) ([]*ent.RpgLotteryPool, int, error) {
	q := r.client.RpgLotteryPool.Query()
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	rows, err := q.Order(ent.Asc(rpglotterypool.FieldSort)).Offset(offset).Limit(limit).All(ctx)
	return rows, total, err
}

// DeleteLotteryPool 删除奖池行。
func (r *RpgRepo) DeleteLotteryPool(ctx context.Context, id int) error {
	return r.client.RpgLotteryPool.DeleteOneID(id).Exec(ctx)
}

// CreateLotteryRecord 写入抽奖记录。
func (r *RpgRepo) CreateLotteryRecord(ctx context.Context, row *ent.RpgUserLotteryRecord) (*ent.RpgUserLotteryRecord, error) {
	b := r.client.RpgUserLotteryRecord.Create().
		SetUID(row.UID).
		SetPoolItemCode(row.PoolItemCode).
		SetItemName(row.ItemName).
		SetRarity(row.Rarity)
	if row.EffectJson != nil {
		b.SetEffectJson(*row.EffectJson)
	}
	return b.Save(ctx)
}

// ListLotteryRecordsByUID 用户抽奖历史。
func (r *RpgRepo) ListLotteryRecordsByUID(ctx context.Context, uid, limit int) ([]*ent.RpgUserLotteryRecord, error) {
	return r.client.RpgUserLotteryRecord.Query().
		Where(rpguserlotteryrecord.UIDEQ(uid)).
		Order(ent.Desc(rpguserlotteryrecord.FieldID)).
		Limit(limit).
		All(ctx)
}

// ListLotteryRecordsAdmin 管理端抽奖记录。
func (r *RpgRepo) ListLotteryRecordsAdmin(ctx context.Context, offset, limit int) ([]*ent.RpgUserLotteryRecord, int, error) {
	q := r.client.RpgUserLotteryRecord.Query()
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	rows, err := q.Order(ent.Desc(rpguserlotteryrecord.FieldID)).Offset(offset).Limit(limit).All(ctx)
	return rows, total, err
}

// --- Pet ---

// ListPetsByUID 用户宠物列表。
func (r *RpgRepo) ListPetsByUID(ctx context.Context, uid int) ([]*ent.RpgUserPet, error) {
	return r.client.RpgUserPet.Query().
		Where(rpguserpet.UIDEQ(uid), rpguserpet.IsDelete(false)).
		Order(ent.Desc(rpguserpet.FieldCreateTime)).
		All(ctx)
}

// FindPetByID 查宠物实例。
func (r *RpgRepo) FindPetByID(ctx context.Context, id int) (*ent.RpgUserPet, error) {
	return r.client.RpgUserPet.Query().Where(rpguserpet.IDEQ(id), rpguserpet.IsDelete(false)).First(ctx)
}

// CreatePet 创建宠物。
func (r *RpgRepo) CreatePet(ctx context.Context, row *ent.RpgUserPet) (*ent.RpgUserPet, error) {
	return r.client.RpgUserPet.Create().
		SetUID(row.UID).
		SetPetCode(row.PetCode).
		SetLevel(row.Level).
		SetExp(row.Exp).
		SetNickname(row.Nickname).
		Save(ctx)
}

// UpdatePet 更新宠物。
func (r *RpgRepo) UpdatePet(ctx context.Context, row *ent.RpgUserPet) (*ent.RpgUserPet, error) {
	return r.client.RpgUserPet.UpdateOneID(row.ID).
		SetNickname(row.Nickname).
		SetLevel(row.Level).
		SetExp(row.Exp).
		Save(ctx)
}

// CountPetsByUID 统计宠物数量。
func (r *RpgRepo) CountPetsByUID(ctx context.Context, uid int) (int, error) {
	return r.client.RpgUserPet.Query().Where(rpguserpet.UIDEQ(uid), rpguserpet.IsDelete(false)).Count(ctx)
}

// --- Buff ---

// ListBuffsByUID 用户 Buff 列表。
func (r *RpgRepo) ListBuffsByUID(ctx context.Context, uid int) ([]*ent.RpgUserBuff, error) {
	return r.client.RpgUserBuff.Query().
		Where(rpguserbuff.UIDEQ(uid), rpguserbuff.IsDelete(false)).
		Order(ent.Asc(rpguserbuff.FieldExpireAt)).
		All(ctx)
}

// FindBuffByID 查 Buff。
func (r *RpgRepo) FindBuffByID(ctx context.Context, id, uid int) (*ent.RpgUserBuff, error) {
	return r.client.RpgUserBuff.Query().
		Where(rpguserbuff.IDEQ(id), rpguserbuff.UIDEQ(uid), rpguserbuff.IsDelete(false)).
		First(ctx)
}

// CreateBuff 创建 Buff。
func (r *RpgRepo) CreateBuff(ctx context.Context, row *ent.RpgUserBuff) (*ent.RpgUserBuff, error) {
	b := r.client.RpgUserBuff.Create().
		SetUID(row.UID).
		SetBuffCode(row.BuffCode).
		SetBuffType(row.BuffType).
		SetName(row.Name).
		SetDescription(row.Description).
		SetValue(row.Value).
		SetExpireAt(row.ExpireAt).
		SetRemainingUses(row.RemainingUses).
		SetIsActive(row.IsActive).
		SetTriggerMode(row.TriggerMode)
	if row.EffectJson != nil {
		b.SetEffectJson(*row.EffectJson)
	}
	return b.Save(ctx)
}

// UpdateBuff 更新 Buff。
func (r *RpgRepo) UpdateBuff(ctx context.Context, row *ent.RpgUserBuff) (*ent.RpgUserBuff, error) {
	up := r.client.RpgUserBuff.UpdateOneID(row.ID).
		SetExpireAt(row.ExpireAt).
		SetRemainingUses(row.RemainingUses).
		SetIsActive(row.IsActive)
	if row.EffectJson != nil {
		up.SetEffectJson(*row.EffectJson)
	}
	return up.Save(ctx)
}

// DeleteExpiredBuffs 删除过期 Buff。
func (r *RpgRepo) DeleteExpiredBuffs(ctx context.Context, uid int, before time.Time) error {
	_, err := r.client.RpgUserBuff.Delete().
		Where(rpguserbuff.UIDEQ(uid), rpguserbuff.ExpireAtLT(before)).
		Exec(ctx)
	return err
}

// --- Achievement ---

// ListAchievementsByUID 用户成就进度。
func (r *RpgRepo) ListAchievementsByUID(ctx context.Context, uid int) ([]*ent.RpgUserAchievement, error) {
	return r.client.RpgUserAchievement.Query().
		Where(rpguserachievement.UIDEQ(uid), rpguserachievement.IsDelete(false)).
		All(ctx)
}

// FindAchievementProgress 查单条成就进度。
func (r *RpgRepo) FindAchievementProgress(ctx context.Context, uid int, code string) (*ent.RpgUserAchievement, error) {
	return r.client.RpgUserAchievement.Query().
		Where(rpguserachievement.UIDEQ(uid), rpguserachievement.AchievementCodeEQ(code)).
		First(ctx)
}

// SaveAchievementProgress 保存成就进度。
func (r *RpgRepo) SaveAchievementProgress(ctx context.Context, row *ent.RpgUserAchievement) (*ent.RpgUserAchievement, error) {
	if row.ID > 0 {
		up := r.client.RpgUserAchievement.UpdateOneID(row.ID).
			SetProgress(row.Progress).
			SetCompleted(row.Completed)
		if row.CompletedAt != nil {
			up.SetCompletedAt(*row.CompletedAt)
		}
		return up.Save(ctx)
	}
	b := r.client.RpgUserAchievement.Create().
		SetUID(row.UID).
		SetAchievementCode(row.AchievementCode).
		SetProgress(row.Progress).
		SetCompleted(row.Completed)
	if row.CompletedAt != nil {
		b.SetCompletedAt(*row.CompletedAt)
	}
	return b.Save(ctx)
}

// ListAchievementConfigs 成就配置列表。
func (r *RpgRepo) ListAchievementConfigs(ctx context.Context) ([]*ent.RpgItemConfig, error) {
	return r.client.RpgItemConfig.Query().
		Where(rpgitemconfig.ItemTypeEQ("achievement"), rpgitemconfig.ActiveEQ(1)).
		Order(ent.Asc(rpgitemconfig.FieldCategory), ent.Asc(rpgitemconfig.FieldSort)).
		All(ctx)
}

// --- Guild ---

// FindGuildByID 查公会。
func (r *RpgRepo) FindGuildByID(ctx context.Context, id int) (*ent.RpgGuild, error) {
	return r.client.RpgGuild.Query().Where(rpgguild.IDEQ(id), rpgguild.IsDelete(false)).First(ctx)
}

// FindGuildByName 按名称查公会。
func (r *RpgRepo) FindGuildByName(ctx context.Context, name string) (*ent.RpgGuild, error) {
	return r.client.RpgGuild.Query().Where(rpgguild.NameEQ(name), rpgguild.IsDelete(false)).First(ctx)
}

// ListGuilds 公会列表。
func (r *RpgRepo) ListGuilds(ctx context.Context, offset, limit int) ([]*ent.RpgGuild, int, error) {
	q := r.client.RpgGuild.Query().Where(rpgguild.IsDelete(false))
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	rows, err := q.Order(ent.Desc(rpgguild.FieldMemberCount)).Offset(offset).Limit(limit).All(ctx)
	return rows, total, err
}

// CreateGuild 创建公会。
func (r *RpgRepo) CreateGuild(ctx context.Context, row *ent.RpgGuild) (*ent.RpgGuild, error) {
	b := r.client.RpgGuild.Create().
		SetName(row.Name).
		SetLeaderUid(row.LeaderUid).
		SetMemberCount(row.MemberCount)
	if row.Announcement != nil {
		b.SetAnnouncement(*row.Announcement)
	}
	return b.Save(ctx)
}

// UpdateGuild 更新公会。
func (r *RpgRepo) UpdateGuild(ctx context.Context, row *ent.RpgGuild) (*ent.RpgGuild, error) {
	up := r.client.RpgGuild.UpdateOneID(row.ID).
		SetMemberCount(row.MemberCount)
	if row.Announcement != nil {
		up.SetAnnouncement(*row.Announcement)
	}
	return up.Save(ctx)
}

// FindGuildMemberByUID 查成员关系。
func (r *RpgRepo) FindGuildMemberByUID(ctx context.Context, uid int) (*ent.RpgUserGuildMember, error) {
	return r.client.RpgUserGuildMember.Query().
		Where(rpguserguildmember.UIDEQ(uid)).
		First(ctx)
}

// ListGuildMembers 公会成员列表。
func (r *RpgRepo) ListGuildMembers(ctx context.Context, guildID int) ([]*ent.RpgUserGuildMember, error) {
	return r.client.RpgUserGuildMember.Query().
		Where(rpguserguildmember.GuildIdEQ(guildID)).
		All(ctx)
}

// CreateGuildMember 加入公会。
func (r *RpgRepo) CreateGuildMember(ctx context.Context, row *ent.RpgUserGuildMember) (*ent.RpgUserGuildMember, error) {
	joinTime := row.JoinTime
	if joinTime.IsZero() {
		joinTime = time.Now()
	}
	return r.client.RpgUserGuildMember.Create().
		SetGuildId(row.GuildId).
		SetUID(row.UID).
		SetRole(row.Role).
		SetJoinTime(joinTime).
		Save(ctx)
}

// DeleteGuildMember 退出公会。
func (r *RpgRepo) DeleteGuildMember(ctx context.Context, id int) error {
	return r.client.RpgUserGuildMember.DeleteOneID(id).Exec(ctx)
}

// DeleteGuild 删除公会及其成员记录。
func (r *RpgRepo) DeleteGuild(ctx context.Context, id int) error {
	return r.WithTx(ctx, func(tx *ent.Tx) error {
		if _, err := tx.RpgUserGuildMember.Delete().Where(rpguserguildmember.GuildIdEQ(id)).Exec(ctx); err != nil {
			return err
		}
		return tx.RpgGuild.DeleteOneID(id).Exec(ctx)
	})
}

// --- Activity ---

// FindActivityByCode 查活动。
func (r *RpgRepo) FindActivityByCode(ctx context.Context, code string) (*ent.RpgActivity, error) {
	return r.client.RpgActivity.Query().Where(rpgactivity.CodeEQ(code)).First(ctx)
}

// ListCurrentActivities 当前有效活动。
func (r *RpgRepo) ListCurrentActivities(ctx context.Context, now time.Time) ([]*ent.RpgActivity, error) {
	return r.client.RpgActivity.Query().
		Where(
			rpgactivity.ActiveEQ(1),
			rpgactivity.StartTimeLTE(now),
			rpgactivity.EndTimeGTE(now),
		).
		Order(ent.Desc(rpgactivity.FieldStartTime)).
		All(ctx)
}

// ListActiveActivitiesStartingBetween 今日开始的活动（active=1）。
func (r *RpgRepo) ListActiveActivitiesStartingBetween(ctx context.Context, from, to time.Time) ([]*ent.RpgActivity, error) {
	return r.client.RpgActivity.Query().
		Where(
			rpgactivity.ActiveEQ(1),
			rpgactivity.StartTimeGTE(from),
			rpgactivity.StartTimeLTE(to),
		).
		All(ctx)
}

// ListActiveActivitiesEndingBetween 今日结束的活动（active=1）。
func (r *RpgRepo) ListActiveActivitiesEndingBetween(ctx context.Context, from, to time.Time) ([]*ent.RpgActivity, error) {
	return r.client.RpgActivity.Query().
		Where(
			rpgactivity.ActiveEQ(1),
			rpgactivity.EndTimeGTE(from),
			rpgactivity.EndTimeLTE(to),
		).
		All(ctx)
}

// CreateActivity 创建活动。
func (r *RpgRepo) CreateActivity(ctx context.Context, row *ent.RpgActivity) (*ent.RpgActivity, error) {
	return r.client.RpgActivity.Create().
		SetCode(row.Code).
		SetName(row.Name).
		SetDescription(row.Description).
		SetActivityType(row.ActivityType).
		SetStartTime(row.StartTime).
		SetEndTime(row.EndTime).
		SetExpBuffRate(row.ExpBuffRate).
		SetPosterUrl(row.PosterUrl).
		SetActive(row.Active).
		Save(ctx)
}

// ListActivitiesAdmin 管理端活动列表。
func (r *RpgRepo) ListActivitiesAdmin(ctx context.Context, offset, limit int) ([]*ent.RpgActivity, int, error) {
	q := r.client.RpgActivity.Query()
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	rows, err := q.Order(ent.Desc(rpgactivity.FieldStartTime)).Offset(offset).Limit(limit).All(ctx)
	return rows, total, err
}

// UpdateActivity 更新活动。
func (r *RpgRepo) UpdateActivity(ctx context.Context, id int, patch map[string]interface{}) error {
	up := r.client.RpgActivity.UpdateOneID(id)
	for k, v := range patch {
		switch k {
		case rpgactivity.FieldName:
			up.SetName(v.(string))
		case rpgactivity.FieldDescription:
			up.SetDescription(v.(string))
		case rpgactivity.FieldActive:
			up.SetActive(v.(int))
		case rpgactivity.FieldExpBuffRate:
			up.SetExpBuffRate(v.(float64))
		}
	}
	_, err := up.Save(ctx)
	return err
}

// DeleteActivity 删除活动。
func (r *RpgRepo) DeleteActivity(ctx context.Context, id int) error {
	return r.client.RpgActivity.DeleteOneID(id).Exec(ctx)
}

// --- Social / Tip ---

// CreateSocialLog 写入社交日志。
func (r *RpgRepo) CreateSocialLog(ctx context.Context, row *ent.RpgUserSocialLog) (*ent.RpgUserSocialLog, error) {
	return r.client.RpgUserSocialLog.Create().
		SetFromUid(row.FromUid).
		SetToUid(row.ToUid).
		SetAction(row.Action).
		SetCostCurrency(row.CostCurrency).
		SetHpDelta(row.HpDelta).
		Save(ctx)
}

// ListSocialLogsAdmin 管理端社交流水。
func (r *RpgRepo) ListSocialLogsAdmin(ctx context.Context, offset, limit int) ([]*ent.RpgUserSocialLog, int, error) {
	q := r.client.RpgUserSocialLog.Query()
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	rows, err := q.Order(ent.Desc(rpgusersociallog.FieldCreateTime)).Offset(offset).Limit(limit).All(ctx)
	return rows, total, err
}

// CreateArticleTip 创建打赏记录。
func (r *RpgRepo) CreateArticleTip(ctx context.Context, row *ent.RpgArticleTip) (*ent.RpgArticleTip, error) {
	return r.client.RpgArticleTip.Create().
		SetUID(row.UID).
		SetArticleId(row.ArticleId).
		SetAuthorUid(row.AuthorUid).
		SetAmount(row.Amount).
		Save(ctx)
}

// ListTipsAdmin 管理端打赏流水。
func (r *RpgRepo) ListTipsAdmin(ctx context.Context, offset, limit int) ([]*ent.RpgArticleTip, int, error) {
	q := r.client.RpgArticleTip.Query()
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	rows, err := q.Order(ent.Desc(rpgarticletip.FieldCreateTime)).Offset(offset).Limit(limit).All(ctx)
	return rows, total, err
}

// --- Level Reward ---

// ListActiveLevelRewards 列出启用等级奖励。
func (r *RpgRepo) ListActiveLevelRewards(ctx context.Context) ([]*ent.RpgLevelReward, error) {
	return r.client.RpgLevelReward.Query().
		Where(rpglevelreward.ActiveEQ(1)).
		Order(ent.Asc(rpglevelreward.FieldSort), ent.Asc(rpglevelreward.FieldLevel)).
		All(ctx)
}

// FindLevelRewardByLevel 按等级查奖励配置。
func (r *RpgRepo) FindLevelRewardByLevel(ctx context.Context, level int) (*ent.RpgLevelReward, error) {
	return r.client.RpgLevelReward.Query().Where(rpglevelreward.LevelEQ(level)).First(ctx)
}

// CreateLevelReward 创建等级奖励配置。
func (r *RpgRepo) CreateLevelReward(ctx context.Context, row *ent.RpgLevelReward) (*ent.RpgLevelReward, error) {
	return r.client.RpgLevelReward.Create().
		SetLevel(row.Level).
		SetAvatarFrame(row.AvatarFrame).
		SetTitle(row.Title).
		SetCurrencyReward(row.CurrencyReward).
		SetActive(row.Active).
		SetSort(row.Sort).
		Save(ctx)
}

// UpdateLevelReward 更新等级奖励。
func (r *RpgRepo) UpdateLevelReward(ctx context.Context, row *ent.RpgLevelReward) (*ent.RpgLevelReward, error) {
	return r.client.RpgLevelReward.UpdateOneID(row.ID).
		SetLevel(row.Level).
		SetAvatarFrame(row.AvatarFrame).
		SetTitle(row.Title).
		SetCurrencyReward(row.CurrencyReward).
		SetActive(row.Active).
		SetSort(row.Sort).
		Save(ctx)
}

// DeleteLevelRewardByID 删除等级奖励。
func (r *RpgRepo) DeleteLevelRewardByID(ctx context.Context, id int) error {
	return r.client.RpgLevelReward.DeleteOneID(id).Exec(ctx)
}

// --- Leaderboard Snapshot ---

// CreateLeaderboardSnapshot 创建排行榜快照。
func (r *RpgRepo) CreateLeaderboardSnapshot(ctx context.Context, row *ent.RpgLeaderboardSnapshot) (*ent.RpgLeaderboardSnapshot, error) {
	return r.client.RpgLeaderboardSnapshot.Create().
		SetUID(row.UID).
		SetScore(row.Score).
		SetRank(row.Rank).
		SetPeriodType(row.PeriodType).
		SetPeriodKey(row.PeriodKey).
		SetScoreType(row.ScoreType).
		Save(ctx)
}

// UpdateLeaderboardSnapshot 更新排行榜快照。
func (r *RpgRepo) UpdateLeaderboardSnapshot(ctx context.Context, row *ent.RpgLeaderboardSnapshot) (*ent.RpgLeaderboardSnapshot, error) {
	return r.client.RpgLeaderboardSnapshot.UpdateOneID(row.ID).
		SetUID(row.UID).
		SetScore(row.Score).
		SetRank(row.Rank).
		SetPeriodType(row.PeriodType).
		SetPeriodKey(row.PeriodKey).
		SetScoreType(row.ScoreType).
		Save(ctx)
}

// DeleteLeaderboardSnapshotByID 软删排行榜快照。
func (r *RpgRepo) DeleteLeaderboardSnapshotByID(ctx context.Context, id int) error {
	return r.client.RpgLeaderboardSnapshot.DeleteOneID(id).Exec(ctx)
}

// ListLeaderboardSnapshots 按周期查排行榜快照。
func (r *RpgRepo) ListLeaderboardSnapshots(ctx context.Context, scoreType, periodType, periodKey string) ([]*ent.RpgLeaderboardSnapshot, error) {
	return r.client.RpgLeaderboardSnapshot.Query().
		Where(
			rpgleaderboardsnapshot.ScoreTypeEQ(scoreType),
			rpgleaderboardsnapshot.PeriodTypeEQ(periodType),
			rpgleaderboardsnapshot.PeriodKeyEQ(periodKey),
			rpgleaderboardsnapshot.IsDelete(false),
		).
		Order(ent.Asc(rpgleaderboardsnapshot.FieldRank)).
		All(ctx)
}

// ListRpgAdmin 管理端 RPG 用户列表。
func (r *RpgRepo) ListRpgAdmin(ctx context.Context, offset, limit int) ([]*ent.Rpg, int, error) {
	q := r.client.Rpg.Query().Where(rpg.IsDelete(false))
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	rows, err := q.Order(ent.Desc(rpg.FieldExp)).Offset(offset).Limit(limit).All(ctx)
	return rows, total, err
}

// FindPayOrderByOutTradeNo 查支付订单。
func (r *RpgRepo) FindPayOrderByOutTradeNo(ctx context.Context, outTradeNo string) (*ent.PayOrder, error) {
	return r.client.PayOrder.Query().Where(payorder.OutTradeNoEQ(outTradeNo), payorder.IsDelete(false)).First(ctx)
}

// CreatePayOrder 创建支付订单。
func (r *RpgRepo) CreatePayOrder(ctx context.Context, row *ent.PayOrder) (*ent.PayOrder, error) {
	b := r.client.PayOrder.Create().
		SetOutTradeNo(row.OutTradeNo).
		SetSubject(row.Subject).
		SetTotalAmount(row.TotalAmount).
		SetStatus(row.Status).
		SetChannel(row.Channel)
	if row.ExtendParams != nil {
		b.SetExtendParams(row.ExtendParams)
	}
	return b.Save(ctx)
}

// UpdatePayOrder 更新支付订单。
func (r *RpgRepo) UpdatePayOrder(ctx context.Context, row *ent.PayOrder) (*ent.PayOrder, error) {
	up := r.client.PayOrder.UpdateOneID(row.ID).
		SetStatus(row.Status).
		SetTradeNo(row.TradeNo)
	if row.ExtendParams != nil {
		up.SetExtendParams(row.ExtendParams)
	}
	return up.Save(ctx)
}
