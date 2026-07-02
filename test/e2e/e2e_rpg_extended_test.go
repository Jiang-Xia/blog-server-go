//go:build e2e

package e2e_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/Jiang-Xia/blog-server-go/test/testutil"
)

func TestE2ERPGQuestsAndAchievements(t *testing.T) {
	c, err := testutil.NewClient("")
	if err != nil {
		t.Fatal(err)
	}
	testutil.RequireServer(t, c)
	ctx := context.Background()
	token := testutil.MustLogin(t, c)

	quests, status, err := c.GET(ctx, "/rpg/my-quests", token)
	testutil.AssertOK(t, "my-quests", quests, status, err)
	var qd struct {
		Daily  []struct{ Code string `json:"code"` } `json:"daily"`
		Weekly []struct{ Code string `json:"code"` } `json:"weekly"`
		Bounty []any `json:"bounty"`
		Special []any `json:"special"`
	}
	if err := testutil.UnmarshalData(quests, &qd); err != nil {
		t.Fatal(err)
	}
	if len(qd.Daily) < 1 || len(qd.Weekly) < 1 {
		t.Fatalf("quest seeds missing: daily=%d weekly=%d", len(qd.Daily), len(qd.Weekly))
	}

	ach, status, err := c.GET(ctx, "/rpg/my-achievements", token)
	testutil.AssertOK(t, "achievements", ach, status, err)
	var achList []struct{ Code string `json:"code"` }
	if err := testutil.UnmarshalData(ach, &achList); err != nil || len(achList) < 10 {
		t.Fatalf("achievements too few: %d", len(achList))
	}

	catalog, status, err := c.GET(ctx, "/rpg/pets/catalog", "")
	testutil.AssertOK(t, "pets catalog", catalog, status, err)
	var pets []struct{ Code string `json:"code"` }
	if err := testutil.UnmarshalData(catalog, &pets); err != nil {
		t.Fatal(err)
	}
	hasCat := false
	for _, p := range pets {
		if p.Code == "pet_cat" {
			hasCat = true
			break
		}
	}
	if !hasCat {
		t.Log("warn: pet_cat not in catalog")
	}

	seasonLb, status, err := c.GET(ctx, "/rpg/leaderboard?type=exp&period=season&limit=5", "")
	testutil.AssertOK(t, "season leaderboard", seasonLb, status, err)
}

func TestE2ERPGPetRenameAndExchange(t *testing.T) {
	c, err := testutil.NewClient("")
	if err != nil {
		t.Fatal(err)
	}
	testutil.RequireServer(t, c)
	ctx := context.Background()
	token := testutil.MustLogin(t, c)

	pets, status, err := c.GET(ctx, "/rpg/pets", token)
	testutil.AssertOK(t, "pets", pets, status, err)
	var petList []struct {
		ID int `json:"id"`
	}
	if err := testutil.UnmarshalData(pets, &petList); err != nil {
		t.Fatal(err)
	}
	if len(petList) > 0 {
		rename, status, err := c.PATCH(ctx, testutil.Pathf("/rpg/pets/%d/rename", petList[0].ID),
			map[string]any{"nickname": "E2E宠物"}, token)
		testutil.AssertOK(t, "pet rename", rename, status, err)
	}

	testutil.AssertBizError(t, c, ctx, http.MethodPost, "/rpg/pets/exchange",
		map[string]any{"petCode": "nonexistent_pet"}, token, 400)

	st, status, err := c.GET(ctx, "/rpg/status", token)
	testutil.AssertOK(t, "status for exchange", st, status, err)
	var statusData struct {
		Currency int `json:"currency"`
	}
	_ = testutil.UnmarshalData(st, &statusData)
	if statusData.Currency >= 50 {
		ex, status, err := c.POST(ctx, "/rpg/pets/exchange", map[string]any{"petCode": "pet_slime"}, token)
		if err != nil || status != 200 {
			t.Fatalf("exchange: http=%d err=%v", status, err)
		}
		if !testutil.IsOK(ex) && ex.Code != 400 {
			t.Fatalf("exchange unexpected code=%d msg=%s", ex.Code, ex.Message)
		}
	} else {
		ex, status, err := c.POST(ctx, "/rpg/pets/exchange", map[string]any{"petCode": "pet_slime"}, token)
		if err != nil || status != 200 || ex.Code != 400 {
			t.Fatalf("exchange insufficient currency want 400, got code=%d", ex.Code)
		}
	}
}

func TestE2ERPGSocialAndGuild(t *testing.T) {
	c, err := testutil.NewClient("")
	if err != nil {
		t.Fatal(err)
	}
	testutil.RequireServer(t, c)
	ctx := context.Background()
	token1 := testutil.MustLogin(t, c)
	token2 := testutil.MustSignToken(t, 2, "18888888889")

	cheer, status, err := c.POST(ctx, "/rpg/social/cheer", map[string]any{"targetUid": 1}, token2)
	if err != nil || status != 200 {
		t.Fatalf("cheer: http=%d err=%v", status, err)
	}
	if !testutil.IsOK(cheer) && cheer.Code != 400 {
		t.Fatalf("cheer code=%d msg=%s", cheer.Code, cheer.Message)
	}

	st2, status, err := c.GET(ctx, "/rpg/status", token2)
	testutil.AssertOK(t, "uid2 status", st2, status, err)
	var uid2 struct{ Currency int `json:"currency"` }
	_ = testutil.UnmarshalData(st2, &uid2)
	if uid2.Currency >= 10 {
		flower, status, err := c.POST(ctx, "/rpg/social/flower", map[string]any{"targetUid": 1}, token2)
		if err != nil || status != 200 {
			t.Fatalf("flower: http=%d err=%v", status, err)
		}
		if !testutil.IsOK(flower) && flower.Code != 400 {
			t.Fatalf("flower code=%d", flower.Code)
		}
	}

	myGuild, status, err := c.GET(ctx, "/rpg/guild/my", token1)
	testutil.AssertOK(t, "guild/my", myGuild, status, err)
	var g1 struct{ ID int `json:"id"` }
	_ = testutil.UnmarshalData(myGuild, &g1)

	g2, status, err := c.GET(ctx, "/rpg/guild/my", token2)
	testutil.AssertOK(t, "guild/my uid2", g2, status, err)
	var g2data struct{ ID int `json:"id"` }
	_ = testutil.UnmarshalData(g2, &g2data)

	if g1.ID > 0 {
		testutil.AssertGETOK(t, c, ctx, "", testutil.Pathf("/rpg/guild/%d", g1.ID))
		if g2data.ID == 0 {
			join, status, err := c.POST(ctx, "/rpg/guild/join", map[string]any{"guildId": g1.ID}, token2)
			if testutil.IsOK(join) {
				leave, status, err := c.POST(ctx, "/rpg/guild/leave", nil, token2)
				testutil.AssertOK(t, "guild leave", leave, status, err)
			} else if join.Code != 400 {
				t.Logf("guild join skipped: code=%d msg=%s", join.Code, join.Message)
			}
			_ = status
			_ = err
		}
	}
}

func TestE2ERGPLotteryFlow(t *testing.T) {
	c, err := testutil.NewClient("")
	if err != nil {
		t.Fatal(err)
	}
	testutil.RequireServer(t, c)
	ctx := context.Background()
	token := testutil.MustLogin(t, c)

	tickets, status, err := c.GET(ctx, "/rpg/lottery/tickets", token)
	testutil.AssertOK(t, "lottery tickets", tickets, status, err)
	var ticketData struct{ Tickets int `json:"tickets"` }
	_ = testutil.UnmarshalData(tickets, &ticketData)

	st, status, err := c.GET(ctx, "/rpg/status", token)
	testutil.AssertOK(t, "status", st, status, err)
	var statusData struct{ Currency int `json:"currency"` }
	_ = testutil.UnmarshalData(st, &statusData)

	if ticketData.Tickets > 0 {
		draw, status, err := c.POST(ctx, "/rpg/lottery/draw", map[string]any{
			"count": 1, "currency": "ticket",
		}, token)
		if err != nil || status != 200 || (!testutil.IsOK(draw) && draw.Code != 400) {
			t.Fatalf("ticket draw: code=%d err=%v", draw.Code, err)
		}
	} else if statusData.Currency >= 10 {
		draw, status, err := c.POST(ctx, "/rpg/lottery/draw", map[string]any{
			"count": 1, "currency": "currency",
		}, token)
		if err != nil || status != 200 || (!testutil.IsOK(draw) && draw.Code != 400) {
			t.Fatalf("currency draw: code=%d err=%v", draw.Code, err)
		}
	} else {
		t.Log("skip lottery draw: no tickets and currency < 10")
	}

	hist, status, err := c.GET(ctx, "/rpg/lottery/history", token)
	testutil.AssertOK(t, "lottery history", hist, status, err)
}
