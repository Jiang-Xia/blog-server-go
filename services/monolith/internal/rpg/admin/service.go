// Package admin RPG 管理端 CRUD 桩实现。
package admin

import (
	"context"

	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	rpgcore "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/core"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/guild"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/inventory"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/lottery"
	rpgpunish "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/punishment"
	rpgrepo "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/repo"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent/rpgitemconfig"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent/rpgquest"
)

// Service 管理端业务。
type Service struct {
	repo         *rpgrepo.RpgRepo
	core         *rpgcore.RpgService
	inventory    *inventory.Service
	lottery      *lottery.Service
	guild        *guild.Service
	punishment   *rpgpunish.PunishmentService
	uploadRoot   string
	staticPrefix string
}

// NewService 构造 Admin Service。
func NewService(
	repo *rpgrepo.RpgRepo,
	core *rpgcore.RpgService,
	inventory *inventory.Service,
	lottery *lottery.Service,
	guild *guild.Service,
	punishment *rpgpunish.PunishmentService,
	uploadRoot, staticPrefix string,
) *Service {
	return &Service{
		repo: repo, core: core, inventory: inventory, lottery: lottery, guild: guild,
		punishment: punishment, uploadRoot: uploadRoot, staticPrefix: staticPrefix,
	}
}

// --- 成就 ---

// ListAchievements 成就配置列表。
func (s *Service) ListAchievements(ctx context.Context, page, pageSize int) (map[string]interface{}, error) {
	return s.listItemConfigsByType(ctx, "achievement", page, pageSize)
}

// GetAchievement 成就详情。
func (s *Service) GetAchievement(ctx context.Context, code string) (*ent.RpgItemConfig, error) {
	return s.repo.FindItemConfigByCode(ctx, code)
}

// --- 任务 ---

// ListQuests 任务列表。
func (s *Service) ListQuests(ctx context.Context, page, pageSize int) (map[string]interface{}, error) {
	offset, limit := paginate(page, pageSize)
	rows, total, err := s.repo.ListQuestsAdmin(ctx, offset, limit)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"list": rows, "total": total}, nil
}

// GetQuest 任务详情。
func (s *Service) GetQuest(ctx context.Context, code string) (*ent.RpgQuest, error) {
	return s.repo.FindQuestByCode(ctx, code)
}

// CreateQuest 创建任务。
func (s *Service) CreateQuest(ctx context.Context, row *ent.RpgQuest) (*ent.RpgQuest, error) {
	return s.repo.CreateQuest(ctx, row)
}

// UpdateQuest 更新任务。
func (s *Service) UpdateQuest(ctx context.Context, id int, patch map[string]interface{}) error {
	return s.repo.UpdateQuest(ctx, id, patch)
}

// DeleteQuest 删除任务。
func (s *Service) DeleteQuest(ctx context.Context, id int) error {
	return s.repo.DeleteQuest(ctx, id)
}

// --- 抽奖 ---

// ListLotteryPool 奖池列表。
func (s *Service) ListLotteryPool(ctx context.Context, page, pageSize int) (map[string]interface{}, error) {
	offset, limit := paginate(page, pageSize)
	rows, total, err := s.repo.ListLotteryPoolAdmin(ctx, offset, limit)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"list": rows, "total": total}, nil
}

// ListLotteryRecords 抽奖记录。
func (s *Service) ListLotteryRecords(ctx context.Context, page, pageSize int) (map[string]interface{}, error) {
	offset, limit := paginate(page, pageSize)
	rows, total, err := s.repo.ListLotteryRecordsAdmin(ctx, offset, limit)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"list": rows, "total": total}, nil
}

// GetLotteryPool 奖池详情。
func (s *Service) GetLotteryPool(ctx context.Context, id int) (*ent.RpgLotteryPool, error) {
	pools, err := s.repo.ListActiveLotteryPool(ctx)
	if err != nil {
		return nil, err
	}
	for _, p := range pools {
		if p.ID == id {
			return p, nil
		}
	}
	return nil, errcode.NotFound
}

// --- 用户 RPG ---

// ListUsers 用户 RPG 列表。
func (s *Service) ListUsers(ctx context.Context, page, pageSize int) (map[string]interface{}, error) {
	offset, limit := paginate(page, pageSize)
	rows, total, err := s.repo.ListRpgAdmin(ctx, offset, limit)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"list": rows, "total": total}, nil
}

// GetUserRpg 单用户 RPG。
func (s *Service) GetUserRpg(ctx context.Context, uid int) (*ent.Rpg, error) {
	return s.core.GetOrCreateRpg(ctx, uid)
}

// RechargeDiamonds 管理员补钻。
func (s *Service) RechargeDiamonds(ctx context.Context, uid, amount int) (int, error) {
	return s.inventory.AdjustCurrency(ctx, uid, amount, "admin_recharge")
}

// DeductDiamonds 管理员扣钻。
func (s *Service) DeductDiamonds(ctx context.Context, uid, amount int) (int, error) {
	return s.inventory.AdjustCurrency(ctx, uid, -amount, "admin_deduct")
}

// --- 物品 ---

// ListItems 物品配置列表。
func (s *Service) ListItems(ctx context.Context, page, pageSize int) (map[string]interface{}, error) {
	offset, limit := paginate(page, pageSize)
	rows, total, err := s.repo.ListItemConfigsAdmin(ctx, offset, limit)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"list": rows, "total": total}, nil
}

// CreateItem 创建物品。
func (s *Service) CreateItem(ctx context.Context, row *ent.RpgItemConfig) (*ent.RpgItemConfig, error) {
	return s.repo.CreateItemConfig(ctx, row)
}

// UpdateItem 更新物品。
func (s *Service) UpdateItem(ctx context.Context, id int, patch map[string]interface{}) error {
	return s.repo.UpdateItemConfig(ctx, id, patch)
}

// DeleteItem 删除物品。
func (s *Service) DeleteItem(ctx context.Context, id int) error {
	return s.repo.DeleteItemConfig(ctx, id)
}

// --- 活动 ---

// ListActivities 活动列表。
func (s *Service) ListActivities(ctx context.Context, page, pageSize int) (map[string]interface{}, error) {
	offset, limit := paginate(page, pageSize)
	rows, total, err := s.repo.ListActivitiesAdmin(ctx, offset, limit)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"list": rows, "total": total}, nil
}

// UpdateActivity 更新活动。
func (s *Service) UpdateActivity(ctx context.Context, id int, patch map[string]interface{}) error {
	return s.repo.UpdateActivity(ctx, id, patch)
}

// DeleteActivity 删除活动。
func (s *Service) DeleteActivity(ctx context.Context, id int) error {
	return s.repo.DeleteActivity(ctx, id)
}

// --- 公会 ---

// ListGuilds 公会列表。
func (s *Service) ListGuilds(ctx context.Context, page, pageSize int) (map[string]interface{}, error) {
	return s.guild.List(ctx, page, pageSize)
}

// --- 流水 ---

// ListSocialLogs 社交流水。
func (s *Service) ListSocialLogs(ctx context.Context, page, pageSize int) (map[string]interface{}, error) {
	offset, limit := paginate(page, pageSize)
	rows, total, err := s.repo.ListSocialLogsAdmin(ctx, offset, limit)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"list": rows, "total": total}, nil
}

// ListTips 打赏流水。
func (s *Service) ListTips(ctx context.Context, page, pageSize int) (map[string]interface{}, error) {
	offset, limit := paginate(page, pageSize)
	rows, total, err := s.repo.ListTipsAdmin(ctx, offset, limit)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"list": rows, "total": total}, nil
}

// GetStats 简易统计桩。
func (s *Service) GetStats(ctx context.Context) (map[string]interface{}, error) {
	_, totalUsers, _ := s.repo.ListRpgAdmin(ctx, 0, 1)
	_, totalQuests, _ := s.repo.ListQuestsAdmin(ctx, 0, 1)
	return map[string]interface{}{
		"totalUsers":  totalUsers,
		"totalQuests": totalQuests,
	}, nil
}

func (s *Service) listItemConfigsByType(ctx context.Context, itemType string, page, pageSize int) (map[string]interface{}, error) {
	rows, err := s.repo.ListItemConfigsByType(ctx, itemType, false)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"list": rows, "total": len(rows)}, nil
}

func paginate(page, pageSize int) (offset, limit int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return (page - 1) * pageSize, pageSize
}

// 避免未使用 import 编译错误（patch 字段常量引用）。
var (
	_ = rpgitemconfig.FieldCode
	_ = rpgquest.FieldCode
)
