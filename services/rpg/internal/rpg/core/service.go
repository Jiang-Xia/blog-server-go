// Package core RPG 主数据业务：用户记录 CRUD 与完整状态聚合。
package core

import (
	"context"

	"github.com/Jiang-Xia/blog-server-go/services/rpg/ent"
	rpgconst "github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/constants"
	rpgrepo "github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/repo"
)

// CosmeticSummary 已解锁装扮摘要（C 端展示）。
type CosmeticSummary struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	Rarity string `json:"rarity"`
}

// FullStatus 用户完整 RPG 状态，对齐 Nest getFullStatus 响应字段。
type FullStatus struct {
	Level                       int               `json:"level"`
	Exp                         int               `json:"exp"`
	LifeValue                   int               `json:"lifeValue"`
	UnlockedAvatarFrames        []CosmeticSummary `json:"unlockedAvatarFrames"`
	UnlockedTitles              []CosmeticSummary `json:"unlockedTitles"`
	EquippedTitle               *string           `json:"equippedTitle"`
	EquippedAvatarFrame         *string           `json:"equippedAvatarFrame"`
	EquippedPetID               *int              `json:"equippedPetId"`
	Currency                    int               `json:"currency"`
	Reputation                  int               `json:"reputation"`
	LotteryPityCounter          int               `json:"lotteryPityCounter"`
	LotteryLegendaryPityCounter int               `json:"lotteryLegendaryPityCounter"`
	LotteryTickets              int               `json:"lotteryTickets"`
	TotalSignDays               int               `json:"totalSignDays"`
	ConsecutiveSignDays         int               `json:"consecutiveSignDays"`
	LastSignDate                interface{}       `json:"lastSignDate"`
	SensitiveHitsCount          int               `json:"sensitiveHitsCount"`
	ZeroLifeCount               int               `json:"zeroLifeCount"`
	BanStartTime                interface{}       `json:"banStartTime"`
	BanEndTime                  interface{}       `json:"banEndTime"`
}

// RpgService RPG 主记录读写与状态聚合。
type RpgService struct {
	repo *rpgrepo.RpgRepo
}

// NewRpgService 构造 RpgService。
func NewRpgService(repo *rpgrepo.RpgRepo) *RpgService {
	return &RpgService{repo: repo}
}

// GetOrCreateRpg 获取或初始化用户 RPG 记录；首次创建时同步初始化装扮行。
func (s *RpgService) GetOrCreateRpg(ctx context.Context, uid int) (*ent.Rpg, error) {
	row, err := s.repo.FindRpgByUID(ctx, uid)
	if err == nil {
		return row, nil
	}
	if !ent.IsNotFound(err) {
		return nil, err
	}
	created, err := s.repo.CreateRpg(ctx, uid)
	if err != nil {
		return nil, err
	}
	_, _ = s.repo.GetOrCreateLoadout(ctx, uid)
	return created, nil
}

// SaveRpg 持久化 RPG 实体变更。
func (s *RpgService) SaveRpg(ctx context.Context, rpg *ent.Rpg) (*ent.Rpg, error) {
	return s.repo.UpdateRpg(ctx, rpg)
}

// FindByUid 按 uid 查询 RPG 记录，不存在返回 nil。
func (s *RpgService) FindByUid(ctx context.Context, uid int) (*ent.Rpg, error) {
	row, err := s.repo.FindRpgByUID(ctx, uid)
	if ent.IsNotFound(err) {
		return nil, nil
	}
	return row, err
}

// GetFullStatus 获取用户完整 RPG 状态（背包钻石/装扮来自 Ent 查询）。
func (s *RpgService) GetFullStatus(ctx context.Context, uid int) (*FullStatus, error) {
	rpg, err := s.GetOrCreateRpg(ctx, uid)
	if err != nil {
		return nil, err
	}

	loadout, err := s.repo.GetOrCreateLoadout(ctx, uid)
	if err != nil {
		return nil, err
	}

	currency := 0
	if inv, err := s.repo.FindInventoryByUIDAndItemCode(ctx, uid, rpgconst.CurrencyItemCode); err == nil {
		currency = inv.Quantity
	} else if !ent.IsNotFound(err) {
		return nil, err
	}

	titles, frames, err := s.cosmeticSummaries(ctx, uid)
	if err != nil {
		return nil, err
	}

	status := &FullStatus{
		Level:                       rpg.Level,
		Exp:                         rpg.Exp,
		LifeValue:                   rpg.LifeValue,
		UnlockedAvatarFrames:        frames,
		UnlockedTitles:              titles,
		EquippedTitle:               loadout.TitleCode,
		EquippedAvatarFrame:         loadout.AvatarFrameCode,
		EquippedPetID:               loadout.PetId,
		Currency:                    currency,
		Reputation:                  rpg.Reputation,
		LotteryPityCounter:          rpg.LotteryPityCounter,
		LotteryLegendaryPityCounter: rpg.LotteryLegendaryPityCounter,
		LotteryTickets:              rpg.LotteryTickets,
		TotalSignDays:               rpg.TotalSignDays,
		ConsecutiveSignDays:         rpg.ConsecutiveSignDays,
		SensitiveHitsCount:          rpg.SensitiveHitsCount,
		ZeroLifeCount:               rpg.ZeroLifeCount,
	}
	if rpg.LastSignDate != nil {
		status.LastSignDate = *rpg.LastSignDate
	}
	if rpg.BanStartTime != nil {
		status.BanStartTime = *rpg.BanStartTime
	}
	if rpg.BanEndTime != nil {
		status.BanEndTime = *rpg.BanEndTime
	}
	return status, nil
}

func (s *RpgService) cosmeticSummaries(ctx context.Context, uid int) (titles, frames []CosmeticSummary, err error) {
	inventory, err := s.repo.ListInventoryByUID(ctx, uid)
	if err != nil {
		return nil, nil, err
	}
	cfgCache := map[string]*ent.RpgItemConfig{}
	for _, inv := range inventory {
		cfg, ok := cfgCache[inv.ItemCode]
		if !ok {
			cfg, err = s.repo.FindItemConfigByCode(ctx, inv.ItemCode)
			if ent.IsNotFound(err) {
				continue
			}
			if err != nil {
				return nil, nil, err
			}
			cfgCache[inv.ItemCode] = cfg
		}
		summary := CosmeticSummary{Code: cfg.Code, Name: cfg.Name, Rarity: cfg.Rarity}
		switch cfg.ItemType {
		case "title":
			titles = append(titles, summary)
		case "avatar_frame":
			frames = append(frames, summary)
		}
	}
	if titles == nil {
		titles = []CosmeticSummary{}
	}
	if frames == nil {
		frames = []CosmeticSummary{}
	}
	return titles, frames, nil
}
