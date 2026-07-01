// Package inventory 背包、装扮槽与钻石余额。
package inventory

import (
	"context"
	"time"

	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	rpgconst "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/constants"
	rpgcore "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/core"
	rpgrepo "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/repo"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/seeds"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent/rpguserinventory"
	"go.uber.org/zap"
)

var legacyCurrencyCodes = map[string]struct{}{
	rpgconst.CurrencyItemCode: {},
	"diamond":                 {}, // 历史种子数据兼容
}

// LoadoutSlot 装扮槽类型。
type LoadoutSlot string

const (
	LoadoutTitle       LoadoutSlot = "title"
	LoadoutAvatarFrame LoadoutSlot = "avatar_frame"
	LoadoutPet         LoadoutSlot = "pet"
)

// Service 背包与货币业务。
type Service struct {
	repo *rpgrepo.RpgRepo
	core *rpgcore.RpgService
	log  *zap.Logger
}

// NewService 构造背包 Service。
func NewService(repo *rpgrepo.RpgRepo, core *rpgcore.RpgService, log *zap.Logger) *Service {
	return &Service{repo: repo, core: core, log: log}
}

// SyncPredefinedSeeds 启动时同步预定义物品种子。
func (s *Service) SyncPredefinedSeeds(ctx context.Context) error {
	return seeds.SyncAllPredefined(ctx, s.repo, s.log)
}

// GetCurrency 查询用户钻石余额。
func (s *Service) GetCurrency(ctx context.Context, uid int) (int, error) {
	inv, err := s.repo.FindInventoryByUIDAndCode(ctx, uid, rpgconst.CurrencyItemCode)
	if ent.IsNotFound(err) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return inv.Quantity, nil
}

// AdjustCurrency 调整钻石余额（事务 + 悲观锁语义：Ent 事务内读写）。
func (s *Service) AdjustCurrency(ctx context.Context, uid, delta int, source string) (int, error) {
	var balance int
	err := s.repo.WithTx(ctx, func(tx *ent.Tx) error {
		inv, err := tx.RpgUserInventory.Query().
			Where(
				rpguserinventory.UIDEQ(uid),
				rpguserinventory.ItemCodeEQ(rpgconst.CurrencyItemCode),
			).
			First(ctx)
		if ent.IsNotFound(err) {
			if delta < 0 {
				return errcode.WithMessage(errcode.InvalidParam, "钻石不足")
			}
			created, err := tx.RpgUserInventory.Create().
				SetUID(uid).
				SetItemCode(rpgconst.CurrencyItemCode).
				SetQuantity(delta).
				SetSource(source).
				SetAcquiredAt(time.Now()).
				Save(ctx)
			if err != nil {
				return err
			}
			balance = created.Quantity
			return nil
		}
		if err != nil {
			return err
		}
		next := inv.Quantity + delta
		if next < 0 {
			return errcode.WithMessage(errcode.InvalidParam, "钻石不足")
		}
		updated, err := tx.RpgUserInventory.UpdateOneID(inv.ID).SetQuantity(next).Save(ctx)
		if err != nil {
			return err
		}
		balance = updated.Quantity
		return nil
	})
	return balance, err
}

// GetInventory 获取用户背包列表（quantity>0）。
func (s *Service) GetInventory(ctx context.Context, uid int) ([]map[string]interface{}, error) {
	rows, err := s.repo.ListInventoryByUID(ctx, uid)
	if err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		cfg, _ := s.repo.FindItemConfigByCode(ctx, row.ItemCode)
		item := map[string]interface{}{
			"itemCode":  row.ItemCode,
			"quantity":  row.Quantity,
			"source":    row.Source,
			"acquiredAt": row.AcquiredAt,
		}
		if cfg != nil {
			item["name"] = cfg.Name
			item["itemType"] = cfg.ItemType
			item["rarity"] = cfg.Rarity
			item["icon"] = cfg.Icon
		}
		out = append(out, item)
	}
	return out, nil
}

// GetLoadout 获取装扮槽配置。
func (s *Service) GetLoadout(ctx context.Context, uid int) (map[string]interface{}, error) {
	loadout, err := s.repo.GetOrCreateLoadout(ctx, uid)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"titleCode":       loadout.TitleCode,
		"avatarFrameCode": loadout.AvatarFrameCode,
		"petId":           loadout.PetId,
	}, nil
}

// GetLoadoutDetail 获取装扮槽详情（含物品名）。
func (s *Service) GetLoadoutDetail(ctx context.Context, uid int) (map[string]interface{}, error) {
	loadout, err := s.repo.GetOrCreateLoadout(ctx, uid)
	if err != nil {
		return nil, err
	}
	result := map[string]interface{}{
		"titleCode":       loadout.TitleCode,
		"avatarFrameCode": loadout.AvatarFrameCode,
		"petId":           loadout.PetId,
	}
	if loadout.TitleCode != nil {
		if cfg, err := s.repo.FindItemConfigByCode(ctx, *loadout.TitleCode); err == nil {
			result["titleName"] = cfg.Name
		}
	}
	if loadout.AvatarFrameCode != nil {
		if cfg, err := s.repo.FindItemConfigByCode(ctx, *loadout.AvatarFrameCode); err == nil {
			result["avatarFrameName"] = cfg.Name
		}
	}
	return result, nil
}

// Equip 穿戴装扮或宠物。
func (s *Service) Equip(ctx context.Context, uid int, slot LoadoutSlot, code string, petID *int) error {
	loadout, err := s.repo.GetOrCreateLoadout(ctx, uid)
	if err != nil {
		return err
	}
	switch slot {
	case LoadoutTitle, LoadoutAvatarFrame:
		if code == "" {
			return errcode.WithMessage(errcode.InvalidParam, "物品编码不能为空")
		}
		if ok, err := s.HasItem(ctx, uid, code); err != nil {
			return err
		} else if !ok {
			return errcode.WithMessage(errcode.NotFound, "未持有该物品")
		}
		if slot == LoadoutTitle {
			loadout.TitleCode = &code
		} else {
			loadout.AvatarFrameCode = &code
		}
	case LoadoutPet:
		if petID == nil || *petID <= 0 {
			return errcode.WithMessage(errcode.InvalidParam, "宠物ID无效")
		}
		pet, err := s.repo.FindPetByID(ctx, *petID)
		if err != nil {
			return errcode.WithMessage(errcode.NotFound, "宠物不存在")
		}
		if pet.UID != uid {
			return errcode.WithMessage(errcode.Forbidden, "无权操作该宠物")
		}
		loadout.PetId = petID
	default:
		return errcode.WithMessage(errcode.InvalidParam, "无效装扮槽")
	}
	_, err = s.repo.UpdateLoadout(ctx, loadout)
	return err
}

// Unequip 卸下装扮槽。
func (s *Service) Unequip(ctx context.Context, uid int, slot LoadoutSlot) error {
	loadout, err := s.repo.GetOrCreateLoadout(ctx, uid)
	if err != nil {
		return err
	}
	switch slot {
	case LoadoutTitle:
		loadout.TitleCode = nil
	case LoadoutAvatarFrame:
		loadout.AvatarFrameCode = nil
	case LoadoutPet:
		loadout.PetId = nil
	default:
		return errcode.WithMessage(errcode.InvalidParam, "无效装扮槽")
	}
	_, err = s.repo.UpdateLoadout(ctx, loadout)
	return err
}

// GetCosmeticSummaries 获取已持有装扮摘要。
func (s *Service) GetCosmeticSummaries(ctx context.Context, uid int, itemTypes []string) ([]rpgcore.CosmeticSummary, error) {
	rows, err := s.repo.ListInventoryByUID(ctx, uid)
	if err != nil {
		return nil, err
	}
	typeSet := map[string]struct{}{}
	for _, t := range itemTypes {
		typeSet[t] = struct{}{}
	}
	out := make([]rpgcore.CosmeticSummary, 0)
	for _, inv := range rows {
		cfg, err := s.repo.FindItemConfigByCode(ctx, inv.ItemCode)
		if err != nil {
			continue
		}
		if len(typeSet) > 0 {
			if _, ok := typeSet[cfg.ItemType]; !ok {
				continue
			}
		}
		out = append(out, rpgcore.CosmeticSummary{Code: cfg.Code, Name: cfg.Name, Rarity: cfg.Rarity})
	}
	return out, nil
}

// GrantItem 发放物品；货币类走 AdjustCurrency。
func (s *Service) GrantItem(ctx context.Context, uid int, itemCode, source string, quantity int) error {
	if _, ok := legacyCurrencyCodes[itemCode]; ok {
		// 旧物品编码 currency/diamond 等映射到通用货币背包。
		_, err := s.AdjustCurrency(ctx, uid, quantity, source)
		return err
	}
	existing, err := s.repo.FindInventoryByUIDAndCode(ctx, uid, itemCode)
	if ent.IsNotFound(err) {
		_, err = s.repo.CreateInventory(ctx, &ent.RpgUserInventory{
			UID:      uid,
			ItemCode: itemCode,
			Quantity: quantity,
			Source:   source,
		})
		return err
	}
	if err != nil {
		return err
	}
	return s.repo.UpdateInventoryQuantity(ctx, existing.ID, existing.Quantity+quantity)
}

// HasItem 是否持有物品（货币查余额）。
func (s *Service) HasItem(ctx context.Context, uid int, itemCode string) (bool, error) {
	if _, ok := legacyCurrencyCodes[itemCode]; ok {
		n, err := s.GetCurrency(ctx, uid)
		return n > 0, err
	}
	inv, err := s.repo.FindInventoryByUIDAndCode(ctx, uid, itemCode)
	if ent.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return inv.Quantity > 0, nil
}

// ConsumeItem 消耗背包物品 1 个。
func (s *Service) ConsumeItem(ctx context.Context, uid int, itemCode string) error {
	inv, err := s.repo.FindInventoryByUIDAndCode(ctx, uid, itemCode)
	if err != nil {
		return errcode.WithMessage(errcode.NotFound, "物品不足")
	}
	if inv.Quantity <= 1 {
		return s.repo.UpdateInventoryQuantity(ctx, inv.ID, 0)
	}
	return s.repo.UpdateInventoryQuantity(ctx, inv.ID, inv.Quantity-1)
}
