// Package quest 每日/周常/特殊任务进度与领奖。
package quest

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	rpgcore "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/core"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/inventory"
	rpglevel "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/level"
	rpgrepo "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/repo"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/util"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
)

// Service 任务业务。
type Service struct {
	repo      *rpgrepo.RpgRepo
	core      *rpgcore.RpgService
	level     *rpglevel.LevelService
	inventory *inventory.Service
	lottery   LotteryTickets
}

// LotteryTickets 抽奖券增减（避免 import cycle）。
type LotteryTickets interface {
	AddTickets(ctx context.Context, uid, amount int) error
}

// NewService 构造任务 Service。
func NewService(
	repo *rpgrepo.RpgRepo,
	core *rpgcore.RpgService,
	level *rpglevel.LevelService,
	inventory *inventory.Service,
	lottery LotteryTickets,
) *Service {
	return &Service{repo: repo, core: core, level: level, inventory: inventory, lottery: lottery}
}

// SetLottery 延迟注入抽奖券服务（打破 quest/lottery 循环依赖）。
func (s *Service) SetLottery(l LotteryTickets) {
	s.lottery = l
}

// ListActiveQuests 获取活跃任务列表。
func (s *Service) ListActiveQuests(ctx context.Context, questType string) ([]*ent.RpgQuest, error) {
	return s.repo.ListActiveQuests(ctx, questType)
}

// GetMyQuests 获取用户任务进度（daily/bounty/weekly/special 四组）。
func (s *Service) GetMyQuests(ctx context.Context, uid int) (map[string][]map[string]interface{}, error) {
	quests, err := s.repo.ListActiveQuests(ctx, "")
	if err != nil {
		return nil, err
	}

	progressMap := map[string]*ent.RpgUserQuestProgress{}
	for _, date := range []time.Time{util.TodayQuestDate(), util.WeekQuestDate(), util.SpecialQuestDate} {
		rows, err := s.repo.ListQuestProgressByUIDAndDate(ctx, uid, date)
		if err != nil {
			return nil, err
		}
		for _, p := range rows {
			progressMap[p.QuestCode] = p
		}
	}

	mapQuest := func(q *ent.RpgQuest) map[string]interface{} {
		prog := progressMap[q.Code]
		item := map[string]interface{}{
			"code":           q.Code,
			"name":           q.Name,
			"description":    q.Description,
			"type":           q.Type,
			"questSubtype":   q.QuestSubtype,
			"targetAction":   q.TargetAction,
			"targetCount":    q.TargetCount,
			"expReward":      q.ExpReward,
			"hpReward":       q.HpReward,
			"currencyReward": q.CurrencyReward,
			"progress":       0,
			"completed":      false,
			"claimed":        false,
		}
		if prog != nil {
			item["progress"] = prog.Progress
			item["completed"] = prog.Completed == 1
			item["claimed"] = prog.Claimed == 1
		}
		return item
	}

	result := map[string][]map[string]interface{}{
		"daily":   {},
		"bounty":  {},
		"weekly":  {},
		"special": {},
	}
	for _, q := range quests {
		mapped := mapQuest(q)
		sub := q.QuestSubtype
		if sub == "" {
			sub = "daily"
		}
		switch {
		case q.Type == "weekly" || sub == "weekly":
			result["weekly"] = append(result["weekly"], mapped)
		case sub == "special":
			result["special"] = append(result["special"], mapped)
		case sub == "bounty":
			result["bounty"] = append(result["bounty"], mapped)
		default:
			result["daily"] = append(result["daily"], mapped)
		}
	}
	return result, nil
}

// TrackProgress 按 targetAction 递增匹配任务进度。
func (s *Service) TrackProgress(ctx context.Context, uid int, action string) error {
	quests, err := s.repo.ListActiveQuests(ctx, "")
	if err != nil {
		return err
	}
	for _, q := range quests {
		if q.TargetAction != action {
			continue
		}
		dateKey := util.QuestDateKey(q.Type, q.QuestSubtype)
		prog, err := s.repo.FindQuestProgress(ctx, uid, q.Code, dateKey)
		if ent.IsNotFound(err) {
			prog = &ent.RpgUserQuestProgress{
				UID:       uid,
				QuestCode: q.Code,
				QuestDate: dateKey,
			}
		} else if err != nil {
			return err
		}
		if prog.Completed == 1 {
			continue
		}
		prog.Progress++
		if prog.Progress >= q.TargetCount {
			prog.Progress = q.TargetCount
			prog.Completed = 1
		}
		if _, err := s.repo.SaveQuestProgress(ctx, prog); err != nil {
			return err
		}
	}
	return nil
}

// ClaimReward 领取任务奖励。
func (s *Service) ClaimReward(ctx context.Context, uid int, questCode string) (map[string]interface{}, error) {
	q, err := s.repo.FindQuestByCode(ctx, questCode)
	if err != nil {
		return nil, errcode.WithMessage(errcode.NotFound, "任务不存在")
	}
	dateKey := util.QuestDateKey(q.Type, q.QuestSubtype)
	prog, err := s.repo.FindQuestProgress(ctx, uid, questCode, dateKey)
	if err != nil {
		return nil, errcode.WithMessage(errcode.InvalidParam, "任务尚未完成")
	}
	if prog.Completed != 1 {
		return nil, errcode.WithMessage(errcode.InvalidParam, "任务尚未完成")
	}
	if prog.Claimed == 1 {
		return nil, errcode.WithMessage(errcode.Conflict, "奖励已领取")
	}

	if q.ExpReward > 0 {
		_, _ = s.level.AddExp(ctx, uid, q.ExpReward, rpgcore.ExpReasonQuest, 0)
	}
	if q.CurrencyReward > 0 {
		_, _ = s.inventory.AdjustCurrency(ctx, uid, q.CurrencyReward, "quest")
	}
	if q.HpReward > 0 {
		rpg, err := s.core.GetOrCreateRpg(ctx, uid)
		if err == nil {
			rpg.LifeValue = min(100, rpg.LifeValue+q.HpReward)
			_, _ = s.core.SaveRpg(ctx, rpg)
		}
	}
	if q.EffectJson != nil && s.lottery != nil {
		effect := util.ParseEffectJSON(q.EffectJson)
		if ticket, ok := effect["ticketReward"].(float64); ok && int(ticket) > 0 {
			_ = s.lottery.AddTickets(ctx, uid, int(ticket))
		}
	}

	prog.Claimed = 1
	if _, err := s.repo.SaveQuestProgress(ctx, prog); err != nil {
		return nil, err
	}
	return map[string]interface{}{"success": true, "questCode": questCode}, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ParseQuestEffect 解析任务 effectJson。
func ParseQuestEffect(raw *string) map[string]interface{} {
	if raw == nil {
		return nil
	}
	var m map[string]interface{}
	_ = json.Unmarshal([]byte(*raw), &m)
	return m
}
