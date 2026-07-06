// Package buff 用户 Buff 授予、激活与经验加成。
package buff

import (
	"context"
	"encoding/json"
	"math/rand"
	"time"

	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	rpgrepo "github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/repo"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/ent"
)

const signInBuffChance = 0.3
const manualExpShelfDays = 7

// Config Buff 配置。
type Config struct {
	Code            string
	Type            string
	Name            string
	Description     string
	Value           float64
	DurationMinutes int
	MaxUses         int
}

// BuffPool 预定义 Buff 池。
var BuffPool = []Config{
	{Code: "exp_boost_small", Type: "exp_boost", Name: "经验微增", Description: "经验获取+20%，激活后持续1小时", Value: 0.2, DurationMinutes: 60, MaxUses: -1},
	{Code: "exp_boost_medium", Type: "exp_boost", Name: "经验大增", Description: "经验获取+50%，激活后持续30分钟", Value: 0.5, DurationMinutes: 30, MaxUses: -1},
	{Code: "hp_regen_double", Type: "hp_regen", Name: "生命加速", Description: "签到HP恢复翻倍", Value: 2.0, DurationMinutes: 1440, MaxUses: 1},
	{Code: "shield_basic", Type: "shield", Name: "护盾", Description: "免疫1次敏感词扣血", Value: 1.0, DurationMinutes: 1440, MaxUses: 1},
	{Code: "lucky_star", Type: "lucky", Name: "幸运星", Description: "签到额外+5经验", Value: 5, DurationMinutes: 1440, MaxUses: 1},
}

// Service Buff 业务。
type Service struct {
	repo *rpgrepo.RpgRepo
}

// NewService 构造 Buff Service。
func NewService(repo *rpgrepo.RpgRepo) *Service {
	return &Service{repo: repo}
}

// GetMyBuffs 获取用户当前有效 Buff 列表。
func (s *Service) GetMyBuffs(ctx context.Context, uid int) ([]map[string]interface{}, error) {
	now := time.Now()
	_ = s.repo.DeleteExpiredBuffs(ctx, uid, now)
	rows, err := s.repo.ListBuffsByUID(ctx, uid)
	if err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, 0)
	for _, b := range rows {
		if b.ExpireAt.Before(now) || b.RemainingUses == 0 {
			continue
		}
		out = append(out, map[string]interface{}{
			"id":            b.ID,
			"buffCode":      b.BuffCode,
			"buffType":      b.BuffType,
			"name":          b.Name,
			"description":   b.Description,
			"value":         b.Value,
			"expireAt":      b.ExpireAt,
			"remainingUses": b.RemainingUses,
			"isActive":      b.IsActive == 1,
			"triggerMode":   b.TriggerMode,
		})
	}
	return out, nil
}

// GrantBuffByCode 按编码授予 Buff。
func (s *Service) GrantBuffByCode(ctx context.Context, uid int, code string) error {
	cfg := findBuffConfig(code)
	if cfg == nil {
		return errcode.WithMessage(errcode.NotFound, "Buff不存在")
	}
	return s.grantBuff(ctx, uid, *cfg)
}

func (s *Service) grantBuff(ctx context.Context, uid int, cfg Config) error {
	now := time.Now()
	isManual := cfg.Type == "exp_boost"
	expire := now.Add(time.Duration(cfg.DurationMinutes) * time.Minute)
	trigger := "auto"
	isActive := 1
	var effect *string
	if isManual {
		expire = now.AddDate(0, 0, manualExpShelfDays)
		trigger = "manual"
		isActive = 0
		meta, _ := json.Marshal(map[string]interface{}{"durationMinutes": cfg.DurationMinutes, "activated": false})
		s := string(meta)
		effect = &s
	}
	uses := cfg.MaxUses
	if uses == 0 {
		uses = 1
	}
	_, err := s.repo.CreateBuff(ctx, &ent.RpgUserBuff{
		UID:           uid,
		BuffCode:      cfg.Code,
		BuffType:      cfg.Type,
		Name:          cfg.Name,
		Description:   cfg.Description,
		Value:         cfg.Value,
		ExpireAt:      expire,
		RemainingUses: uses,
		IsActive:      isActive,
		TriggerMode:   trigger,
		EffectJson:    effect,
	})
	return err
}

// ActivateBuff 手动激活经验 Buff。
func (s *Service) ActivateBuff(ctx context.Context, uid, buffID int) error {
	b, err := s.repo.FindBuffByID(ctx, buffID, uid)
	if err != nil {
		return errcode.WithMessage(errcode.NotFound, "Buff不存在")
	}
	if b.TriggerMode != "manual" {
		return errcode.WithMessage(errcode.InvalidParam, "该Buff无需手动激活")
	}
	if b.IsActive == 1 {
		return errcode.WithMessage(errcode.Conflict, "Buff已激活")
	}
	meta := map[string]interface{}{}
	if b.EffectJson != nil {
		_ = json.Unmarshal([]byte(*b.EffectJson), &meta)
	}
	durationMin := 60
	if v, ok := meta["durationMinutes"].(float64); ok {
		durationMin = int(v)
	}
	b.ExpireAt = time.Now().Add(time.Duration(durationMin) * time.Minute)
	b.IsActive = 1
	activated, _ := json.Marshal(map[string]interface{}{"durationMinutes": durationMin, "activated": true})
	effectStr := string(activated)
	b.EffectJson = &effectStr
	_, err = s.repo.UpdateBuff(ctx, b)
	return err
}

// DeactivateBuff 停用经验 Buff。
func (s *Service) DeactivateBuff(ctx context.Context, uid, buffID int) error {
	b, err := s.repo.FindBuffByID(ctx, buffID, uid)
	if err != nil {
		return errcode.WithMessage(errcode.NotFound, "Buff不存在")
	}
	if b.IsActive != 1 {
		return errcode.WithMessage(errcode.InvalidParam, "Buff未激活")
	}
	b.IsActive = 0
	_, err = s.repo.UpdateBuff(ctx, b)
	return err
}

// ApplyExpBoost 应用已激活的经验 Buff 加成。
func (s *Service) ApplyExpBoost(ctx context.Context, uid, amount int) (int, error) {
	if amount <= 0 {
		return amount, nil
	}
	buffs, err := s.GetMyBuffs(ctx, uid)
	if err != nil {
		return amount, err
	}
	multiplier := 1.0
	for _, b := range buffs {
		if b["buffType"] == "exp_boost" && b["isActive"] == true {
			if v, ok := b["value"].(float64); ok {
				multiplier += v
			}
		}
	}
	return int(float64(amount) * multiplier), nil
}

// TriggerSignInBuff 签到随机 Buff。
func (s *Service) TriggerSignInBuff(ctx context.Context, uid int) (*ent.RpgUserBuff, error) {
	if rand.Float64() > signInBuffChance {
		return nil, nil
	}
	cfg := BuffPool[rand.Intn(len(BuffPool))]
	if err := s.grantBuff(ctx, uid, cfg); err != nil {
		return nil, err
	}
	rows, err := s.repo.ListBuffsByUID(ctx, uid)
	if err != nil || len(rows) == 0 {
		return nil, err
	}
	return rows[len(rows)-1], nil
}

func findBuffConfig(code string) *Config {
	for i := range BuffPool {
		if BuffPool[i].Code == code {
			return &BuffPool[i]
		}
	}
	return nil
}

// HasShield 检查并消耗一次护盾 Buff（免疫敏感词扣血，对齐 Nest buffService.hasShield）。
func (s *Service) HasShield(ctx context.Context, uid int) (bool, error) {
	now := time.Now()
	_ = s.repo.DeleteExpiredBuffs(ctx, uid, now)
	rows, err := s.repo.ListBuffsByUID(ctx, uid)
	if err != nil {
		return false, err
	}
	for _, b := range rows {
		if !shieldEligible(b, now) {
			continue
		}
		if err := s.consumeUse(ctx, b); err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

// shieldEligible 判断 Buff 是否为可用护盾（便于单测）。
func shieldEligible(b *ent.RpgUserBuff, now time.Time) bool {
	if b == nil || b.BuffType != "shield" {
		return false
	}
	if b.ExpireAt.Before(now) || b.RemainingUses == 0 {
		return false
	}
	if b.TriggerMode == "manual" && b.IsActive != 1 {
		return false
	}
	return true
}

func (s *Service) consumeUse(ctx context.Context, b *ent.RpgUserBuff) error {
	if b.RemainingUses <= 0 {
		return nil
	}
	b.RemainingUses--
	if b.RemainingUses <= 0 {
		return s.repo.DeleteBuffByID(ctx, b.ID)
	}
	_, err := s.repo.UpdateBuff(ctx, b)
	return err
}
