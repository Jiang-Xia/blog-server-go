//go:build integration

package integration_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/Jiang-Xia/blog-server-go/test/testutil"
)

func TestIntegrationBlogExtended(t *testing.T) {
	c := newClient(t)
	ctx := context.Background()
	token := testutil.MustLogin(t, c)
	articleID := testutil.FirstArticleID(t, c, ctx, token)

	testutil.AssertGETOK(t, c, ctx, "",
		fmt.Sprintf("/article/info?id=%d", articleID),
		fmt.Sprintf("/article/related?id=%d&limit=3", articleID),
		"/reply/findAll?articleId="+fmt.Sprint(articleID),
		"/resources/daily-img",
		"/resources/weather?city=北京",
	)

	resp, status, err := c.POST(ctx, "/article/views", map[string]any{"articleId": articleID}, "")
	testutil.AssertOK(t, "article/views", resp, status, err)

	testutil.AssertGETOK(t, c, ctx, token,
		fmt.Sprintf("/like/check?articleId=%d", articleID),
		fmt.Sprintf("/collect/check?articleId=%d", articleID),
		fmt.Sprintf("/collect/count?articleId=%d", articleID),
		"/like/my-ids",
		"/collect/list?page=1&pageSize=10",
		"/comment/my-list?page=1&pageSize=5",
		"/comment/on-my-articles?page=1&pageSize=5",
		"/reply/my-list?page=1&pageSize=5",
		"/article/my-list?page=1&pageSize=5",
		"/article/author-stats",
	)

	resp, status, err = c.GET(ctx, "/comment/admin?page=1&pageSize=5", token)
	testutil.AssertOK(t, "comment/admin", resp, status, err)

	testutil.AssertGETOK(t, c, ctx, "",
		"/user/public/1/collects?page=1&pageSize=5",
		"/user/public/1/likes?page=1&pageSize=5",
		"/rpg/public/status/batch?uids=1,2",
	)
}

func TestIntegrationRPGExtendedRead(t *testing.T) {
	c := newClient(t)
	ctx := context.Background()
	token1 := testutil.MustSignToken(t, 1, "18888888888")
	token2 := testutil.MustSignToken(t, 2, "18888888889")

	testutil.AssertGETOK(t, c, ctx, "",
		"/rpg/leaderboard?type=exp&period=week&limit=5",
		"/rpg/leaderboard?type=reputation&period=week&limit=5",
		"/rpg/leaderboard?type=exp&period=season&limit=5",
		"/rpg/weather-buff?city=北京",
	)

	testutil.AssertGETOK(t, c, ctx, token1,
		"/rpg/hit-records?page=1&pageSize=5",
		"/rpg/pets",
		"/rpg/guild/my",
		"/rpg/recharge/status",
	)

	myGuild, status, err := c.GET(ctx, "/rpg/guild/my", token1)
	testutil.AssertOK(t, "guild/my uid1", myGuild, status, err)
	var guild struct {
		ID int `json:"id"`
	}
	if err := testutil.UnmarshalData(myGuild, &guild); err == nil && guild.ID > 0 {
		testutil.AssertGETOK(t, c, ctx, "",
			fmt.Sprintf("/rpg/guild/%d", guild.ID),
		)
	}

	testutil.AssertGETOK(t, c, ctx, token2, "/rpg/guild/my")
}

func TestIntegrationRPGUnauthorizedWrite(t *testing.T) {
	c := newClient(t)
	ctx := context.Background()

	testutil.AssertUnauthorizedPOST(t, c, ctx, "/rpg/lottery/draw", map[string]any{"count": 1})
	testutil.AssertUnauthorizedPOST(t, c, ctx, "/rpg/pets/exchange", map[string]any{"petCode": "pet_slime"})
	testutil.AssertUnauthorizedPOST(t, c, ctx, "/rpg/activities/share-poster", nil)
	testutil.AssertUnauthorizedPOST(t, c, ctx, "/rpg/sign", nil)
	testutil.AssertUnauthorizedPOST(t, c, ctx, "/comment/create", map[string]any{
		"articleId": 1, "content": "integration-unauth",
	})
}

func TestIntegrationRPGAdminRead(t *testing.T) {
	c := newClient(t)
	ctx := context.Background()
	token := testutil.MustLogin(t, c)

	testutil.AssertGETOK(t, c, ctx, token,
		"/admin/rpg/achievements?page=1&pageSize=5",
		"/admin/rpg/quests?page=1&pageSize=5",
		"/admin/rpg/lottery/pool?page=1&pageSize=5",
		"/admin/rpg/lottery/records?page=1&pageSize=5",
		"/admin/rpg/items?page=1&pageSize=5",
		"/admin/rpg/activities?page=1&pageSize=5",
		"/admin/rpg/guilds?page=1&pageSize=5",
		"/admin/rpg/users?page=1&pageSize=5",
		"/admin/rpg/tips?page=1&pageSize=5",
		"/admin/rpg/social-logs?page=1&pageSize=5",
		"/admin/rpg/users/1",
	)
}

func TestIntegrationPayAndUser(t *testing.T) {
	c := newClient(t)
	ctx := context.Background()
	token := testutil.MustLogin(t, c)

	resp, status, err := c.GET(ctx, "/pay/order/list?page=1&pageSize=10", token)
	testutil.AssertOK(t, "pay/order/list", resp, status, err)

	testutil.AssertBizError(t, c, ctx, http.MethodGet, "/pay/order/query", nil, token, 400)

	resp, status, err = c.POST(ctx, "/user/list", map[string]any{
		"page": 1, "pageSize": 5,
	}, "")
	testutil.AssertOK(t, "user/list", resp, status, err)

	resp, status, err = c.GET(ctx, "/privilege?page=1&pageSize=10", token)
	testutil.AssertOK(t, "privilege list", resp, status, err)
}

func TestIntegrationRPGInvalidExchange(t *testing.T) {
	c := newClient(t)
	ctx := context.Background()
	token := testutil.MustSignToken(t, 1, "18888888888")

	testutil.AssertBizError(t, c, ctx, http.MethodPost, "/rpg/pets/exchange",
		map[string]any{"petCode": "invalid_pet_xyz"}, token, 400)
}
