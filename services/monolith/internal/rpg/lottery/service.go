// Package lottery 抽奖池、保底与记录。
package lottery

import (
	"context"
	"encoding/json"
	"math/rand"
	"time"

	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	rpgconst "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/constants"
	rpgcore "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/core"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/inventory"
	rpglevel "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/level"
	rpgrepo "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/repo"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/util"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
)

const drawTicketCost = 1

// Service 抽奖业务。
type Service struct {
	repo      *rpgrepo.RpgRepo
	core      *rpgcore.RpgService
	inventory *inventory.Service
	level     *rpglevel.LevelService
	buff      BuffGranter
	quest     QuestTracker
}

// BuffGranter 授予 Buff（避免 cycle）。
type BuffGranter interface {
	GrantBuffByCode(ctx context.Context, uid int, buffCode string) error
}

// QuestTracker 任务进度（避免 cycle）。
type QuestTracker interface {
	TrackProgress(ctx context.Context, uid int, action string) error
}

// NewService 构造抽奖 Service。
func NewService(
	repo *rpgrepo.RpgRepo,
	core *rpgcore.RpgService,
	inventory *inventory.Service,
	level *rpglevel.LevelService,
	buff BuffGranter,
	quest QuestTracker,
) *Service {
	return &Service{repo: repo, core: core, inventory: inventory, level: level, buff: buff, quest: quest}
}

// GetPool 获取奖池（含物品配置）。
func (s *Service) GetPool(ctx context.Context) ([]map[string]interface{}, error) {
	pool, err := s.repo.ListActiveLotteryPool(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, 0, len(pool))
	for _, p := range pool {
		cfg, _ := s.repo.FindItemConfigByCode(ctx, p.ItemCode)
		item := map[string]interface{}{
			"itemCode":    p.ItemCode,
			"probability": p.Probability,
			"rarity":      p.Rarity,
			"sort":        p.Sort,
		}
		if cfg != nil {
			item["name"] = cfg.Name
			item["description"] = cfg.Description
			item["icon"] = cfg.Icon
			item["itemType"] = cfg.ItemType
		}
		out = append(out, item)
	}
	return out, nil
}

// GetTickets 查询抽奖券数量。
func (s *Service) GetTickets(ctx context.Context, uid int) (int, error) {
	rpg, err := s.core.FindByUid(ctx, uid)
	if err != nil || rpg == nil {
		return 0, err
	}
	return rpg.LotteryTickets, nil
}

// AddTickets 增加抽奖券。
func (s *Service) AddTickets(ctx context.Context, uid, amount int) error {
	rpg, err := s.core.GetOrCreateRpg(ctx, uid)
	if err != nil {
		return err
	}
	rpg.LotteryTickets += amount
	_, err = s.core.SaveRpg(ctx, rpg)
	return err
}

// Draw 执行单次抽奖；useTicket=true 消耗券，否则扣钻石。
func (s *Service) Draw(ctx context.Context, uid int, useTicket bool) (map[string]interface{}, error) {
	rpg, err := s.core.GetOrCreateRpg(ctx, uid)
	if err != nil {
		return nil, err
	}
	if useTicket {
		if rpg.LotteryTickets < drawTicketCost {
			return nil, errcode.WithMessage(errcode.InvalidParam, "抽奖券不足")
		}
		rpg.LotteryTickets -= drawTicketCost
	} else {
		if _, err := s.inventory.AdjustCurrency(ctx, uid, -rpgconst.Economy.LotteryCurrencyCost, "lottery"); err != nil {
			return nil, err
		}
	}

	pool, err := s.repo.ListActiveLotteryPool(ctx)
	if err != nil || len(pool) == 0 {
		return nil, errcode.WithMessage(errcode.InternalError, "奖池为空")
	}

	picked := s.pickWithPity(pool, rpg)
	if picked == nil {
		return nil, errcode.WithMessage(errcode.InternalError, "抽奖失败")
	}

	cfg, _ := s.repo.FindItemConfigByCode(ctx, picked.ItemCode)
	itemName := picked.ItemCode
	if cfg != nil {
		itemName = cfg.Name
	}
	effectStr := ""
	if cfg != nil && cfg.EffectJson != nil {
		effectStr = *cfg.EffectJson
	}
	_, _ = s.repo.CreateLotteryRecord(ctx, &ent.RpgUserLotteryRecord{
		UID:          uid,
		PoolItemCode: picked.ItemCode,
		ItemName:     itemName,
		Rarity:       picked.Rarity,
		EffectJson:   strPtr(effectStr),
	})

	if err := s.applyPrize(ctx, uid, cfg); err != nil {
		return nil, err
	}

	if picked.Rarity == "epic" || picked.Rarity == "legendary" {
		if picked.Rarity == "legendary" {
			rpg.LotteryLegendaryPityCounter = 0
		}
		rpg.LotteryPityCounter = 0
	} else {
		rpg.LotteryPityCounter++
		rpg.LotteryLegendaryPityCounter++
	}
	_, _ = s.core.SaveRpg(ctx, rpg)

	if s.quest != nil {
		_ = s.quest.TrackProgress(ctx, uid, "lottery_draw")
	}

	return map[string]interface{}{
		"itemCode": picked.ItemCode,
		"name":     itemName,
		"rarity":   picked.Rarity,
	}, nil
}

// GetHistory 抽奖历史。
func (s *Service) GetHistory(ctx context.Context, uid, limit int) ([]map[string]interface{}, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := s.repo.ListLotteryRecordsByUID(ctx, uid, limit)
	if err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		out = append(out, map[string]interface{}{
			"itemCode":   r.PoolItemCode,
			"itemName":   r.ItemName,
			"rarity":     r.Rarity,
			"createTime": r.CreateTime,
		})
	}
	return out, nil
}

func (s *Service) pickWithPity(pool []*ent.RpgLotteryPool, rpg *ent.Rpg) *ent.RpgLotteryPool {
	if rpg.LotteryLegendaryPityCounter+1 >= rpgconst.Economy.LotteryLegendaryPityThreshold {
		if p := firstByRarity(pool, "legendary"); p != nil {
			return p
		}
	}
	if rpg.LotteryPityCounter+1 >= rpgconst.Economy.LotteryEpicPityThreshold {
		if p := firstByRarity(pool, "epic"); p != nil {
			return p
		}
	}
	return weightedPick(pool)
}

func firstByRarity(pool []*ent.RpgLotteryPool, rarity string) *ent.RpgLotteryPool {
	for _, p := range pool {
		if p.Rarity == rarity {
			return p
		}
	}
	return nil
}

func weightedPick(pool []*ent.RpgLotteryPool) *ent.RpgLotteryPool {
	var total float64
	for _, p := range pool {
		total += p.Probability
	}
	if total <= 0 {
		return pool[0]
	}
	r := rand.Float64() * total
	var acc float64
	for _, p := range pool {
		acc += p.Probability
		if r <= acc {
			return p
		}
	}
	return pool[len(pool)-1]
}

func (s *Service) applyPrize(ctx context.Context, uid int, cfg *ent.RpgItemConfig) error {
	if cfg == nil {
		return nil
	}
	effect := util.ParseEffectJSON(cfg.EffectJson)
	switch effect["grantType"] {
	case "exp":
		amount := intFromEffect(effect["amount"])
		_, _ = s.level.AddExp(ctx, uid, amount, rpgcore.ExpReasonLottery, 0)
	case "ticket":
		_ = s.AddTickets(ctx, uid, intFromEffect(effect["amount"]))
	case "buff":
		if code, ok := effect["buffCode"].(string); ok && s.buff != nil {
			_ = s.buff.GrantBuffByCode(ctx, uid, code)
		}
	case "pet":
		if code, ok := effect["petCode"].(string); ok {
			_ = s.inventory.GrantItem(ctx, uid, code, "lottery", 1)
		}
	default:
		_ = s.inventory.GrantItem(ctx, uid, cfg.Code, "lottery", 1)
	}
	return nil
}

func intFromEffect(v interface{}) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case json.Number:
		i, _ := n.Int64()
		return int(i)
	default:
		return 0
	}
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
