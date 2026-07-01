// Package profile 公开 RPG 主页与批量状态。
package profile

import (
	"context"
	"strconv"
	"strings"

	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/achievement"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/inventory"
	rpgcore "github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/core"
	rpgrepo "github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/repo"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/internal/userport"
)

// Service 公开主页业务。
type Service struct {
	users        userport.UserReader
	repo         *rpgrepo.RpgRepo
	core         *rpgcore.RpgService
	inventory    *inventory.Service
	achievements *achievement.Service
}

// NewService 构造 Profile Service。
func NewService(
	users userport.UserReader,
	repo *rpgrepo.RpgRepo,
	core *rpgcore.RpgService,
	inventory *inventory.Service,
	achievements *achievement.Service,
) *Service {
	return &Service{users: users, repo: repo, core: core, inventory: inventory, achievements: achievements}
}

// GetPublicProfile 聚合公开主页数据。
func (s *Service) GetPublicProfile(ctx context.Context, uid int) (map[string]interface{}, error) {
	user, err := s.users.FindByID(ctx, uid)
	if err != nil || user == nil || user.IsDelete || user.Status == "0" {
		return nil, errcode.WithMessage(errcode.NotFound, "用户不存在")
	}
	rpg, _ := s.core.FindByUid(ctx, uid)
	loadout, _ := s.inventory.GetLoadoutDetail(ctx, uid)
	achievements, _ := s.achievements.GetMyAchievements(ctx, uid)
	completed := make([]map[string]interface{}, 0)
	for _, a := range achievements {
		if done, ok := a["completed"].(bool); ok && done {
			completed = append(completed, a)
			if len(completed) >= 6 {
				break
			}
		}
	}
	level := 1
	reputation := 0
	if rpg != nil {
		level = rpg.Level
		reputation = rpg.Reputation
	}
	return map[string]interface{}{
		"uid":                  uid,
		"nickname":             user.Nickname,
		"username":             user.Username,
		"avatar":               user.Avatar,
		"intro":                user.Intro,
		"createTime":           user.CreateTime,
		"level":                level,
		"reputation":           reputation,
		"loadout":              loadout,
		"completedAchievements": completed,
	}, nil
}

// GetPublicRpgStatus 单用户 RPG 公开状态。
func (s *Service) GetPublicRpgStatus(ctx context.Context, uid int) (map[string]interface{}, error) {
	rpg, _ := s.core.FindByUid(ctx, uid)
	loadout, _ := s.inventory.GetLoadoutDetail(ctx, uid)
	if rpg == nil {
		return map[string]interface{}{"level": 1, "reputation": 0, "loadout": loadout}, nil
	}
	return map[string]interface{}{
		"level":         rpg.Level,
		"reputation":    rpg.Reputation,
		"totalSignDays": rpg.TotalSignDays,
		"loadout":       loadout,
	}, nil
}

// GetPublicRpgLevelsBatch 批量查作者等级徽章。
func (s *Service) GetPublicRpgLevelsBatch(ctx context.Context, uids []int) (map[int]map[string]int, error) {
	if len(uids) == 0 {
		return map[int]map[string]int{}, nil
	}
	if len(uids) > 100 {
		uids = uids[:100]
	}
	rpgs, err := s.repo.ListRpgByUIDs(ctx, uids)
	if err != nil {
		return nil, err
	}
	levelByUID := map[int]int{}
	for _, r := range rpgs {
		levelByUID[r.UID] = r.Level
	}
	out := map[int]map[string]int{}
	for _, uid := range uids {
		level := levelByUID[uid]
		if level == 0 {
			level = 1
		}
		out[uid] = map[string]int{"level": level}
	}
	return out, nil
}

// ParsePublicRpgUids 解析逗号分隔 uid 字符串。
func ParsePublicRpgUids(raw string) []int {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	seen := map[int]struct{}{}
	out := make([]int, 0)
	for _, part := range strings.Split(raw, ",") {
		id, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}
