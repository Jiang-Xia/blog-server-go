package buff

import (
	"testing"
	"time"

	"github.com/Jiang-Xia/blog-server-go/services/rpg/ent"
)

func TestShieldEligible(t *testing.T) {
	now := time.Now()
	future := now.Add(time.Hour)
	past := now.Add(-time.Hour)

	cases := []struct {
		name string
		b    *ent.RpgUserBuff
		want bool
	}{
		{
			name: "active auto shield",
			b:    &ent.RpgUserBuff{BuffType: "shield", ExpireAt: future, RemainingUses: 1, TriggerMode: "auto", IsActive: 1},
			want: true,
		},
		{
			name: "expired shield",
			b:    &ent.RpgUserBuff{BuffType: "shield", ExpireAt: past, RemainingUses: 1, TriggerMode: "auto", IsActive: 1},
			want: false,
		},
		{
			name: "no uses",
			b:    &ent.RpgUserBuff{BuffType: "shield", ExpireAt: future, RemainingUses: 0, TriggerMode: "auto", IsActive: 1},
			want: false,
		},
		{
			name: "manual inactive",
			b:    &ent.RpgUserBuff{BuffType: "shield", ExpireAt: future, RemainingUses: 1, TriggerMode: "manual", IsActive: 0},
			want: false,
		},
		{
			name: "wrong type",
			b:    &ent.RpgUserBuff{BuffType: "lucky", ExpireAt: future, RemainingUses: 1, TriggerMode: "auto", IsActive: 1},
			want: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := shieldEligible(tc.b, now); got != tc.want {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}
