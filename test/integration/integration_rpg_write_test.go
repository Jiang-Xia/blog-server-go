//go:build integration

package integration_test

import (
	"context"
	"testing"

	"github.com/Jiang-Xia/blog-server-go/test/testutil"
)

func TestIntegrationRPGLoadoutAndBuff(t *testing.T) {
	c := newClient(t)
	ctx := context.Background()
	token := testutil.MustSignToken(t, 1, "18888888888")

	inv, status, err := c.GET(ctx, "/rpg/inventory", token)
	testutil.AssertOK(t, "inventory", inv, status, err)
	var invData struct {
		Items []struct {
			ItemCode string `json:"itemCode"`
			Config   struct {
				ItemType string `json:"itemType"`
			} `json:"config"`
		} `json:"items"`
	}
	if err := testutil.UnmarshalData(inv, &invData); err != nil {
		t.Fatal(err)
	}

	for _, item := range invData.Items {
		if item.Config.ItemType == "title" && item.ItemCode != "" {
			equip, status, err := c.POST(ctx, "/rpg/loadout/equip", map[string]any{
				"slot": "title", "itemCode": item.ItemCode,
			}, token)
			testutil.AssertOK(t, "loadout equip title", equip, status, err)
			unequip, status, err := c.POST(ctx, "/rpg/loadout/unequip", map[string]any{"slot": "title"}, token)
			testutil.AssertOK(t, "loadout unequip title", unequip, status, err)
			break
		}
	}

	buffs, status, err := c.GET(ctx, "/rpg/my-buffs", token)
	testutil.AssertOK(t, "my-buffs", buffs, status, err)
	var buffList []struct {
		ID          int    `json:"id"`
		TriggerMode string `json:"triggerMode"`
	}
	if err := testutil.UnmarshalData(buffs, &buffList); err != nil {
		return
	}
	for _, b := range buffList {
		if b.TriggerMode == "manual" {
			act, status, err := c.POST(ctx, testutil.Pathf("/rpg/buff/%d/activate", b.ID), nil, token)
			testutil.AssertOK(t, "buff activate", act, status, err)
			deact, status, err := c.POST(ctx, testutil.Pathf("/rpg/buff/%d/deactivate", b.ID), nil, token)
			testutil.AssertOK(t, "buff deactivate", deact, status, err)
			break
		}
	}
}

func TestIntegrationRPGQuestsStructure(t *testing.T) {
	c := newClient(t)
	ctx := context.Background()
	token := testutil.MustSignToken(t, 1, "18888888888")

	resp, status, err := c.GET(ctx, "/rpg/my-quests", token)
	testutil.AssertOK(t, "my-quests", resp, status, err)
	var quests struct {
		Daily   []any `json:"daily"`
		Bounty  []any `json:"bounty"`
		Special []any `json:"special"`
		Weekly  []any `json:"weekly"`
	}
	if err := testutil.UnmarshalData(resp, &quests); err != nil {
		t.Fatal(err)
	}
	if quests.Daily == nil || quests.Weekly == nil {
		t.Fatalf("my-quests missing groups: %+v", quests)
	}

	ach, status, err := c.GET(ctx, "/rpg/my-achievements", token)
	testutil.AssertOK(t, "my-achievements", ach, status, err)
	var achList []map[string]any
	if err := testutil.UnmarshalData(ach, &achList); err != nil || len(achList) == 0 {
		t.Fatalf("achievements empty")
	}
}
