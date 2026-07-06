// Package punishment RPG 禁言状态查询、敏感词惩罚链与 admin 解封。
package punishment

import (
	"context"
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

type shieldChecker interface {
	HasShield(ctx context.Context, uid int) (bool, error)
}

type punishmentNotifier interface {
	NotifyShieldUsed(ctx context.Context, uid int)
	NotifyLifeChange(ctx context.Context, uid, lifeDeducted, currentLife int)
	NotifyBanStatus(ctx context.Context, uid int, banned bool, banEndTime *time.Time, banReason *string)
}

// PunishmentService 禁言判定与敏感词惩罚链。
type PunishmentService struct {
	rpg    *rpgcore.RpgService
	buff   shieldChecker
	notify punishmentNotifier
}

// NewPunishmentService 构造 PunishmentService；buff/notify 可 nil（单测或未装配 WS 时）。
func NewPunishmentService(rpg *rpgcore.RpgService, buff shieldChecker, notify punishmentNotifier) *PunishmentService {
	return &PunishmentService{rpg: rpg, buff: buff, notify: notify}
}

// OnSensitiveWordHit 敏感词命中：扣 HP、护盾、累计/归零禁言与 WS 通知（对齐 Nest）。
func (s *PunishmentService) OnSensitiveWordHit(ctx context.Context, uid, hpPenalty int) error {
	if uid <= 0 {
		return nil
	}
	rpg, err := s.rpg.GetOrCreateRpg(ctx, uid)
	if err != nil {
		return err
	}

	hasShield := false
	if s.buff != nil {
		hasShield, err = s.buff.HasShield(ctx, uid)
		if err != nil {
			hasShield = false
		}
	}

	now := time.Now()
	result, shieldUsed := applySensitiveWordHit(rpg, hpPenalty, hasShield, now)
	if _, err := s.rpg.SaveRpg(ctx, rpg); err != nil {
		return err
	}

	if s.notify == nil {
		return nil
	}
	if shieldUsed {
		s.notify.NotifyShieldUsed(ctx, uid)
		return nil
	}
	if result.LifeDeducted > 0 {
		s.notify.NotifyLifeChange(ctx, uid, result.LifeDeducted, result.CurrentLife)
	}
	if result.Banned {
		s.notify.NotifyBanStatus(ctx, uid, true, result.BanEndTime, result.BanReason)
	}
	return nil
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

// AdminUnban 管理员手动解禁：清除禁言时间并推送 banStatus WS。
func (s *PunishmentService) AdminUnban(ctx context.Context, uid int) (map[string]interface{}, error) {
	if uid <= 0 {
		return nil, errcode.WithMessage(errcode.InvalidParam, "无效用户")
	}
	rpg, err := s.rpg.FindByUid(ctx, uid)
	if err != nil {
		return nil, err
	}
	if rpg == nil {
		return nil, errcode.WithMessage(errcode.NotFound, "用户RPG数据不存在")
	}
	rpg.BanStartTime = nil
	rpg.BanEndTime = nil
	if _, err := s.rpg.SaveRpg(ctx, rpg); err != nil {
		return nil, err
	}
	if s.notify != nil {
		s.notify.NotifyBanStatus(ctx, uid, false, nil, nil)
	}
	return map[string]interface{}{"uid": uid, "unbanned": true}, nil
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
	return errcode.WithMessage(errcode.Forbidden, "您已被禁言，剩余%d小时（解禁时间: %s）",
		hours, status.BanEndTime.Format("2006-01-02 15:04"))
}
