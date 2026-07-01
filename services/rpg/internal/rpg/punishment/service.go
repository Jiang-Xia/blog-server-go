// Package punishment RPG 禁言状态查询与断言。
package punishment

import (
	"context"
	"fmt"
	"time"

	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	rpgcore "github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/core"
)

// BanStatus 禁言状态。
type BanStatus struct {
	Banned      bool       `json:"banned"`
	BanEndTime  *time.Time `json:"banEndTime"`
	RemainingMs int64      `json:"remainingMs"`
}

// PunishmentService 禁言判定（敏感词惩罚完整逻辑在后续迭代接入）。
type PunishmentService struct {
	rpg *rpgcore.RpgService
}

// NewPunishmentService 构造 PunishmentService。
func NewPunishmentService(rpg *rpgcore.RpgService) *PunishmentService {
	return &PunishmentService{rpg: rpg}
}

// GetBanStatus 判断用户是否处于禁言状态；过期则自动清除禁言字段。
func (s *PunishmentService) GetBanStatus(ctx context.Context, uid int) (*BanStatus, error) {
	if uid <= 0 {
		return &BanStatus{}, nil
	}
	rpg, err := s.rpg.FindByUid(ctx, uid)
	if err != nil {
		return nil, err
	}
	if rpg == nil || rpg.BanEndTime == nil {
		return &BanStatus{}, nil
	}

	now := time.Now()
	if now.Before(*rpg.BanEndTime) {
		remaining := rpg.BanEndTime.Sub(now).Milliseconds()
		return &BanStatus{
			Banned:      true,
			BanEndTime:  rpg.BanEndTime,
			RemainingMs: remaining,
		}, nil
	}

	rpg.BanStartTime = nil
	rpg.BanEndTime = nil
	if _, err := s.rpg.SaveRpg(ctx, rpg); err != nil {
		return nil, err
	}
	return &BanStatus{}, nil
}

// AssertNotBanned 检查禁言，被禁则返回 Forbidden 业务错误。
func (s *PunishmentService) AssertNotBanned(ctx context.Context, uid int) error {
	status, err := s.GetBanStatus(ctx, uid)
	if err != nil {
		return err
	}
	if !status.Banned || status.BanEndTime == nil {
		return nil
	}
	hours := (status.RemainingMs + int64(time.Hour/time.Millisecond) - 1) / int64(time.Hour/time.Millisecond)
	msg := fmt.Sprintf("您已被禁言，剩余%d小时（解禁时间: %s）",
		hours, status.BanEndTime.Format("2006-01-02 15:04"))
	return errcode.WithMessage(errcode.Forbidden, msg)
}
