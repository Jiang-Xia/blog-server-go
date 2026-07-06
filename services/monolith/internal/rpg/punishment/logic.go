package punishment

import (
	"fmt"
	"time"

	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
)

// PunishmentResult 敏感词惩罚结果，对齐 Nest PunishmentResult。
type PunishmentResult struct {
	LifeDeducted int
	CurrentLife  int
	Banned       bool
	BanEndTime   *time.Time
	BanReason    *string
	HitCount     int
}

// applySensitiveWordHit 在内存中应用敏感词惩罚规则（便于单测固定 clock）。
func applySensitiveWordHit(rpg *ent.Rpg, hpPenalty int, hasShield bool, now time.Time) (PunishmentResult, bool) {
	if hpPenalty <= 0 {
		hpPenalty = LifeDeductPerHit
	}

	rpg.SensitiveHitsCount++

	if hasShield {
		return PunishmentResult{
			CurrentLife: rpg.LifeValue,
			HitCount:    rpg.SensitiveHitsCount,
		}, true
	}

	beforeLife := rpg.LifeValue
	if rpg.LifeValue > hpPenalty {
		rpg.LifeValue -= hpPenalty
	} else {
		rpg.LifeValue = 0
	}
	lifeDeducted := beforeLife - rpg.LifeValue

	banned := false
	var banEndTime *time.Time
	var banReason *string

	if rpg.SensitiveHitsCount > 0 && rpg.SensitiveHitsCount%HitCountBanThreshold == 0 {
		end := now.Add(hitCountBanDuration)
		banEndTime = &end
		rpg.BanStartTime = ptrTime(now)
		rpg.BanEndTime = &end
		banned = true
		reason := fmt.Sprintf("累计命中敏感词%d次，禁言%d小时", rpg.SensitiveHitsCount, HitCountBanHours)
		banReason = &reason
	}

	if rpg.LifeValue <= 0 {
		rpg.ZeroLifeCount++
		rpg.LifeValue = MaxLifeValue

		if rpg.ZeroLifeCount >= ZeroLifePermaBanCount {
			end := now.Add(zeroLifePermaBan)
			banEndTime = &end
			rpg.BanStartTime = ptrTime(now)
			rpg.BanEndTime = &end
			banned = true
			reason := fmt.Sprintf("生命值连续%d次归零，正式禁言30天", ZeroLifePermaBanCount)
			banReason = &reason
		} else if !banned {
			end := now.Add(zeroLifeTempBan)
			banEndTime = &end
			rpg.BanStartTime = ptrTime(now)
			rpg.BanEndTime = &end
			banned = true
			reason := fmt.Sprintf("生命值归零，临时禁言%d小时", ZeroLifeTempBanHours)
			banReason = &reason
		}
	}

	return PunishmentResult{
		LifeDeducted: lifeDeducted,
		CurrentLife:  rpg.LifeValue,
		Banned:       banned,
		BanEndTime:   banEndTime,
		BanReason:    banReason,
		HitCount:     rpg.SensitiveHitsCount,
	}, false
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
