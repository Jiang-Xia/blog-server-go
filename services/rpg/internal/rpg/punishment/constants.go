package punishment

import "time"

// 对齐 Nest punishment.service.ts 常量。
const (
	LifeDeductPerHit       = 20
	MaxLifeValue           = 100
	HitCountBanThreshold   = 5
	HitCountBanHours       = 72
	ZeroLifePermaBanCount  = 3
	ZeroLifeTempBanHours   = 24
	zeroLifePermaBanDays   = 30
)

var (
	hitCountBanDuration = HitCountBanHours * time.Hour
	zeroLifeTempBan     = ZeroLifeTempBanHours * time.Hour
	zeroLifePermaBan    = zeroLifePermaBanDays * 24 * time.Hour
)
