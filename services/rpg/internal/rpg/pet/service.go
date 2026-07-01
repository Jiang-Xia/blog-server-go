// Package pet 宠物孵化、兑换与出战加成。
package pet

import (
	"context"

	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/inventory"
	rpgrepo "github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/repo"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/seeds"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/util"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/ent"
	"go.uber.org/zap"
)

// Service 宠物业务。
type Service struct {
	repo      *rpgrepo.RpgRepo
	inventory *inventory.Service
	quest     QuestTracker
	log       *zap.Logger
}

// QuestTracker 任务/成就追踪接口。
type QuestTracker interface {
	TrackProgress(ctx context.Context, uid int, action string) error
}

// NewService 构造宠物 Service。
func NewService(repo *rpgrepo.RpgRepo, inventory *inventory.Service, quest QuestTracker, log *zap.Logger) *Service {
	return &Service{repo: repo, inventory: inventory, quest: quest, log: log}
}

// SyncPredefinedPets 启动时同步宠物种子（委托 seeds 包）。
func (s *Service) SyncPredefinedPets(ctx context.Context) error {
	petSeeds := []seeds.ItemConfigSeed{
		{Code: "pet_slime", ItemType: "pet", Name: "史莱姆", Rarity: "common", Sort: 1, Active: 1, Effect: map[string]interface{}{"expBoost": 0.05, "currencyCost": 50}},
		{Code: "pet_fox", ItemType: "pet", Name: "灵狐", Rarity: "rare", Sort: 2, Active: 1, Effect: map[string]interface{}{"expBoost": 0.08, "currencyCost": 120}},
		{Code: "pet_cat", ItemType: "pet", Name: "灵猫", Rarity: "common", Sort: 4, Active: 1, Effect: map[string]interface{}{"expBoost": 0.06, "currencyCost": 80}},
	}
	return seeds.UpsertItemConfigSeeds(ctx, s.repo, petSeeds, s.log)
}

// GetCatalog 可兑换宠物目录。
func (s *Service) GetCatalog(ctx context.Context) ([]map[string]interface{}, error) {
	items, err := s.repo.ListItemConfigsByType(ctx, "pet", true)
	if err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		out = append(out, map[string]interface{}{
			"code":        item.Code,
			"name":        item.Name,
			"description": item.Description,
			"rarity":      item.Rarity,
			"icon":        item.Icon,
			"effectJson":  util.ParseEffectJSON(item.EffectJson),
		})
	}
	return out, nil
}

// ListPets 用户宠物列表。
func (s *Service) ListPets(ctx context.Context, uid int) ([]map[string]interface{}, error) {
	pets, err := s.repo.ListPetsByUID(ctx, uid)
	if err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, 0, len(pets))
	for _, p := range pets {
		cfg, _ := s.repo.FindItemConfigByCode(ctx, p.PetCode)
		item := map[string]interface{}{
			"id":       p.ID,
			"petCode":  p.PetCode,
			"nickname": p.Nickname,
			"level":    p.Level,
			"exp":      p.Exp,
		}
		if cfg != nil {
			item["name"] = cfg.Name
			item["config"] = map[string]interface{}{
				"code":   cfg.Code,
				"name":   cfg.Name,
				"rarity": cfg.Rarity,
			}
		}
		out = append(out, item)
	}
	return out, nil
}

// Summon 孵化宠物（消耗宠物蛋或直接 pet 类型物品）。
func (s *Service) Summon(ctx context.Context, uid int, itemCode string) (map[string]interface{}, error) {
	cfg, err := s.repo.FindItemConfigByCode(ctx, itemCode)
	if err != nil {
		return nil, errcode.WithMessage(errcode.NotFound, "物品不存在")
	}
	effect := util.ParseEffectJSON(cfg.EffectJson)
	petCode := itemCode
	if cfg.ItemType == "consumable" && effect["grantType"] == "pet" {
		if code, ok := effect["petCode"].(string); ok {
			petCode = code
		}
		if ok, err := s.inventory.HasItem(ctx, uid, itemCode); err != nil {
			return nil, err
		} else if !ok {
			return nil, errcode.WithMessage(errcode.InvalidParam, "未持有该宠物蛋")
		}
		if err := s.inventory.ConsumeItem(ctx, uid, itemCode); err != nil {
			return nil, err
		}
	} else if cfg.ItemType != "pet" {
		return nil, errcode.WithMessage(errcode.InvalidParam, "无法孵化该物品")
	}
	pet, err := s.repo.CreatePet(ctx, &ent.RpgUserPet{UID: uid, PetCode: petCode, Nickname: cfg.Name})
	if err != nil {
		return nil, err
	}
	if s.quest != nil {
		_ = s.quest.TrackProgress(ctx, uid, "pet_hatch")
	}
	return map[string]interface{}{"id": pet.ID, "petCode": pet.PetCode}, nil
}

// Exchange 钻石兑换宠物。
func (s *Service) Exchange(ctx context.Context, uid int, petCode string) (map[string]interface{}, error) {
	cfg, err := s.repo.FindItemConfigByCode(ctx, petCode)
	if err != nil || cfg.ItemType != "pet" {
		return nil, errcode.WithMessage(errcode.NotFound, "宠物不存在")
	}
	effect := util.ParseEffectJSON(cfg.EffectJson)
	cost := intFromFloat(effect["currencyCost"])
	if cost <= 0 {
		return nil, errcode.WithMessage(errcode.InvalidParam, "该宠物不可兑换")
	}
	if _, err := s.inventory.AdjustCurrency(ctx, uid, -cost, "pet_exchange"); err != nil {
		return nil, err
	}
	return s.Summon(ctx, uid, petCode)
}

// Rename 重命名宠物。
func (s *Service) Rename(ctx context.Context, uid, petID int, nickname string) error {
	pet, err := s.repo.FindPetByID(ctx, petID)
	if err != nil {
		return errcode.WithMessage(errcode.NotFound, "宠物不存在")
	}
	if pet.UID != uid {
		return errcode.WithMessage(errcode.Forbidden, "无权操作")
	}
	pet.Nickname = nickname
	_, err = s.repo.UpdatePet(ctx, pet)
	return err
}

// GetEquippedPetExpBoost 出战宠物经验加成倍率（0~0.15）。
func (s *Service) GetEquippedPetExpBoost(ctx context.Context, uid int) (float64, error) {
	loadout, err := s.repo.GetOrCreateLoadout(ctx, uid)
	if err != nil || loadout.PetId == nil {
		return 0, err
	}
	pet, err := s.repo.FindPetByID(ctx, *loadout.PetId)
	if err != nil {
		return 0, nil
	}
	cfg, err := s.repo.FindItemConfigByCode(ctx, pet.PetCode)
	if err != nil {
		return 0, nil
	}
	effect := util.ParseEffectJSON(cfg.EffectJson)
	return floatFromEffect(effect["expBoost"]), nil
}

func intFromFloat(v interface{}) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	default:
		return 0
	}
}

func floatFromEffect(v interface{}) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	default:
		return 0
	}
}
