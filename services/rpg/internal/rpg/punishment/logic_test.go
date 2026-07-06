package punishment

import (
	"testing"
	"time"

	"github.com/Jiang-Xia/blog-server-go/services/rpg/ent"
)

func fixedNow() time.Time {
	return time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
}

func newTestRpg(life, hits, zeroLife int) *ent.Rpg {
	return &ent.Rpg{LifeValue: life, SensitiveHitsCount: hits, ZeroLifeCount: zeroLife}
}

func TestApplySensitiveWordHitDeductsLife(t *testing.T) {
	rpg := newTestRpg(100, 0, 0)
	result, shieldUsed := applySensitiveWordHit(rpg, 20, false, fixedNow())
	if shieldUsed {
		t.Fatal("expected no shield")
	}
	if result.LifeDeducted != 20 || rpg.LifeValue != 80 {
		t.Fatalf("life=%d deducted=%d", rpg.LifeValue, result.LifeDeducted)
	}
	if rpg.SensitiveHitsCount != 1 {
		t.Fatalf("hits=%d", rpg.SensitiveHitsCount)
	}
}

func TestApplySensitiveWordHitShieldSkipsDeduct(t *testing.T) {
	rpg := newTestRpg(100, 2, 0)
	result, shieldUsed := applySensitiveWordHit(rpg, 20, true, fixedNow())
	if !shieldUsed {
		t.Fatal("expected shield used")
	}
	if result.LifeDeducted != 0 || rpg.LifeValue != 100 {
		t.Fatalf("life should stay 100, got %d", rpg.LifeValue)
	}
	if rpg.SensitiveHitsCount != 3 {
		t.Fatalf("hits should increment, got %d", rpg.SensitiveHitsCount)
	}
}

func TestApplySensitiveWordHitCountBan(t *testing.T) {
	rpg := newTestRpg(100, 4, 0)
	now := fixedNow()
	result, _ := applySensitiveWordHit(rpg, 20, false, now)
	if !result.Banned || rpg.BanEndTime == nil {
		t.Fatal("expected 72h ban at 5th hit")
	}
	wantEnd := now.Add(hitCountBanDuration)
	if !rpg.BanEndTime.Equal(wantEnd) {
		t.Fatalf("ban end=%v want=%v", rpg.BanEndTime, wantEnd)
	}
	if rpg.SensitiveHitsCount != 5 {
		t.Fatalf("hits=%d", rpg.SensitiveHitsCount)
	}
}

func TestApplySensitiveWordHitZeroLifeTempBan(t *testing.T) {
	rpg := newTestRpg(15, 1, 0)
	now := fixedNow()
	result, _ := applySensitiveWordHit(rpg, 20, false, now)
	if !result.Banned {
		t.Fatal("expected temp ban")
	}
	if rpg.LifeValue != MaxLifeValue {
		t.Fatalf("life should reset to %d, got %d", MaxLifeValue, rpg.LifeValue)
	}
	if rpg.ZeroLifeCount != 1 {
		t.Fatalf("zeroLife=%d", rpg.ZeroLifeCount)
	}
	wantEnd := now.Add(zeroLifeTempBan)
	if rpg.BanEndTime == nil || !rpg.BanEndTime.Equal(wantEnd) {
		t.Fatalf("ban end=%v want=%v", rpg.BanEndTime, wantEnd)
	}
}

func TestApplySensitiveWordHitZeroLifePermaBan(t *testing.T) {
	rpg := newTestRpg(10, 1, 2)
	now := fixedNow()
	result, _ := applySensitiveWordHit(rpg, 20, false, now)
	if !result.Banned {
		t.Fatal("expected perma ban")
	}
	if rpg.ZeroLifeCount != 3 {
		t.Fatalf("zeroLife=%d", rpg.ZeroLifeCount)
	}
	wantEnd := now.Add(zeroLifePermaBan)
	if rpg.BanEndTime == nil || !rpg.BanEndTime.Equal(wantEnd) {
		t.Fatalf("ban end=%v want=%v", rpg.BanEndTime, wantEnd)
	}
}

func TestApplySensitiveWordHitDefaultPenalty(t *testing.T) {
	rpg := newTestRpg(100, 0, 0)
	result, _ := applySensitiveWordHit(rpg, 0, false, fixedNow())
	if result.LifeDeducted != LifeDeductPerHit {
		t.Fatalf("default penalty=%d want=%d", result.LifeDeducted, LifeDeductPerHit)
	}
}
