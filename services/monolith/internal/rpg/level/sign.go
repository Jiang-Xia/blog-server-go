package level

import (
	"context"
	"fmt"
	"time"

	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	rpgcore "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/core"
)

const (
	baseSignInExp     = 10
	signInHPRecovery  = 5
	maxLifeValue      = 100
)

type consecutiveBonusTier struct {
	days     int
	bonusExp int
	label    string
}

var consecutiveBonus = []consecutiveBonusTier{
	{days: 3, bonusExp: 5, label: "连续3天"},
	{days: 7, bonusExp: 15, label: "连续7天"},
	{days: 14, bonusExp: 25, label: "连续14天"},
	{days: 30, bonusExp: 50, label: "连续30天"},
}

// SignInResult 签到结果。
type SignInResult struct {
	Success             bool           `json:"success"`
	Message             string         `json:"message"`
	Exp                 int            `json:"exp"`
	Level               int            `json:"level"`
	TotalSignDays       int            `json:"totalSignDays"`
	ConsecutiveSignDays int            `json:"consecutiveSignDays"`
	LifeRecovered       int            `json:"lifeRecovered"`
	BonusExp            int            `json:"bonusExp"`
	BonusLabel          string         `json:"bonusLabel"`
	LevelUp             *rpgcore.LevelUpResult `json:"levelUp"`
}

// SignInfo 签到信息查询结果。
type SignInfo struct {
	SignedToday         bool   `json:"signedToday"`
	TotalSignDays       int    `json:"totalSignDays"`
	ConsecutiveSignDays int    `json:"consecutiveSignDays"`
	LastSignDate        string `json:"lastSignDate"`
	NextBonusAt         *int   `json:"nextBonusAt"`
}

// SignService 每日签到：基础经验、连续奖励、HP 恢复。
type SignService struct {
	rpg   *rpgcore.RpgService
	level *LevelService
}

// NewSignService 构造 SignService。
func NewSignService(rpg *rpgcore.RpgService, level *LevelService) *SignService {
	return &SignService{rpg: rpg, level: level}
}

// SignIn 用户每日签到；防重复基于 lastSignDate 自然日判断。
func (s *SignService) SignIn(ctx context.Context, uid int) (*SignInResult, error) {
	rpg, err := s.rpg.GetOrCreateRpg(ctx, uid)
	if err != nil {
		return nil, err
	}

	today := time.Now().Format("2006-01-02")
	lastSign := formatSignDate(rpg.LastSignDate)
	if lastSign == today {
		return nil, errcode.WithMessage(errcode.InvalidParam, "今日已签到，请明天再来！")
	}

	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	if lastSign == yesterday {
		rpg.ConsecutiveSignDays++
	} else {
		rpg.ConsecutiveSignDays = 1
	}

	bonusExp, bonusLabel := consecutiveBonusFor(rpg.ConsecutiveSignDays)

	hpBefore := rpg.LifeValue
	rpg.LifeValue = minInt(maxLifeValue, rpg.LifeValue+signInHPRecovery)
	lifeRecovered := rpg.LifeValue - hpBefore

	now := time.Now()
	todayTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	rpg.LastSignDate = &todayTime
	rpg.TotalSignDays++
	if _, err := s.rpg.SaveRpg(ctx, rpg); err != nil {
		return nil, err
	}

	totalExp := baseSignInExp + bonusExp
	levelUp, err := s.level.AddExp(ctx, uid, totalExp, rpgcore.ExpReasonSignIn, 0)
	if err != nil {
		return nil, err
	}

	updated, err := s.rpg.GetOrCreateRpg(ctx, uid)
	if err != nil {
		return nil, err
	}

	message := fmt.Sprintf("签到成功！获得%d点经验值，恢复%d点生命值", baseSignInExp, lifeRecovered)
	if bonusExp > 0 {
		message = fmt.Sprintf("签到成功！获得%d+%d点经验（%s），恢复%d点生命值", baseSignInExp, bonusExp, bonusLabel, lifeRecovered)
	}

	return &SignInResult{
		Success:             true,
		Message:             message,
		Exp:                 updated.Exp,
		Level:               updated.Level,
		TotalSignDays:       updated.TotalSignDays,
		ConsecutiveSignDays: updated.ConsecutiveSignDays,
		LifeRecovered:       lifeRecovered,
		BonusExp:            bonusExp,
		BonusLabel:          bonusLabel,
		LevelUp:             levelUp,
	}, nil
}

// GetSignInfo 获取签到信息（今日是否已签、累计/连续天数、下一档连续奖励）。
func (s *SignService) GetSignInfo(ctx context.Context, uid int) (*SignInfo, error) {
	rpg, err := s.rpg.GetOrCreateRpg(ctx, uid)
	if err != nil {
		return nil, err
	}

	today := time.Now().Format("2006-01-02")
	lastSign := formatSignDate(rpg.LastSignDate)

	consecutive := 0
	if lastSign == today {
		consecutive = rpg.ConsecutiveSignDays
	}

	var nextBonus *int
	for _, tier := range consecutiveBonus {
		if tier.days > consecutive {
			d := tier.days
			nextBonus = &d
			break
		}
	}

	info := &SignInfo{
		SignedToday:         lastSign == today,
		TotalSignDays:       rpg.TotalSignDays,
		ConsecutiveSignDays: rpg.ConsecutiveSignDays,
		NextBonusAt:         nextBonus,
	}
	if lastSign != "" {
		info.LastSignDate = lastSign
	}
	return info, nil
}

func formatSignDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02")
}

func consecutiveBonusFor(days int) (bonusExp int, label string) {
	for _, tier := range consecutiveBonus {
		if days >= tier.days {
			bonusExp = tier.bonusExp
			label = tier.label
		}
	}
	return bonusExp, label
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
