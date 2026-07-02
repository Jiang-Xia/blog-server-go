//go:build integration

package integration_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/Jiang-Xia/blog-server-go/test/testutil"
)

func newClient(t *testing.T) *testutil.Client {
	t.Helper()
	c, err := testutil.NewClient("")
	if err != nil {
		t.Fatal(err)
	}
	testutil.RequireServer(t, c)
	return c
}

func TestIntegrationHealthAndCaptcha(t *testing.T) {
	c := newClient(t)
	ctx := context.Background()

	resp, status, err := c.GET(ctx, "/health", "")
	testutil.AssertOK(t, "health", resp, status, err)

	resp, status, err = c.GET(ctx, "/pub/stats", "")
	testutil.AssertOK(t, "pub/stats", resp, status, err)
	var stats map[string]any
	if err := testutil.UnmarshalData(resp, &stats); err != nil {
		t.Fatal(err)
	}
	if _, ok := stats["articleCount"]; !ok {
		t.Fatalf("pub/stats missing articleCount: %v", stats)
	}

	resp, status, err = c.GET(ctx, "/user/authCode", "")
	testutil.AssertOK(t, "authCode", resp, status, err)
	var cap struct {
		CaptchaBase64 string `json:"captchaBase64"`
	}
	if err := testutil.UnmarshalData(resp, &cap); err != nil || cap.CaptchaBase64 == "" {
		t.Fatalf("authCode missing captchaBase64")
	}
}

func TestIntegrationLoginAndJWT(t *testing.T) {
	c := newClient(t)
	ctx := context.Background()

	token, err := c.Login(ctx)
	token = testutil.SkipUnlessLogin(t, c, token, err)
	resp, status, err := c.GET(ctx, "/user/info", token)
	testutil.AssertOK(t, "user/info authed", resp, status, err)

	bad, status, err := c.GET(ctx, "/user/info", "invalid.token")
	if err != nil {
		t.Fatal(err)
	}
	if status != http.StatusOK || !testutil.IsUnauthorized(bad) {
		t.Fatalf("invalid token want code=401, got code=%d", bad.Code)
	}

	refresh, status, err := c.GET(ctx, "/user/refresh?token=invalid-token", "")
	if err != nil || status == http.StatusTooManyRequests {
		t.Fatalf("refresh should not be throttled: status=%d err=%v", status, err)
	}
	if testutil.IsOK(refresh) {
		t.Fatal("invalid refresh should not succeed")
	}
}

func TestIntegrationBlogPublicAPIs(t *testing.T) {
	c := newClient(t)
	ctx := context.Background()

	resp, status, err := c.POST(ctx, "/article/list", map[string]any{
		"page": 1, "pageSize": 10, "client": true,
	}, "")
	testutil.AssertOK(t, "article/list", resp, status, err)
	var list struct {
		Pagination map[string]any `json:"pagination"`
	}
	if err := testutil.UnmarshalData(resp, &list); err != nil || list.Pagination == nil {
		t.Fatalf("article list missing pagination")
	}

	for _, tc := range []struct {
		name string
		path string
	}{
		{"category", "/category?isDelete=true"},
		{"tag", "/tag?isDelete=true"},
		{"archives", "/article/archives"},
		{"statistics", "/article/statistics"},
		{"msgboard", "/msgboard?page=1&pageSize=10"},
		{"link", "/link?client=1"},
		{"register-avatars", "/resources/register-avatars"},
		{"comment", "/comment/findAll?page=1&pageSize=5"},
	} {
		resp, status, err := c.GET(ctx, tc.path, "")
		testutil.AssertOK(t, tc.name, resp, status, err)
	}
}

func TestIntegrationRPGPublicAndAuthed(t *testing.T) {
	c := newClient(t)
	ctx := context.Background()

	token, err := testutil.SignToken(1, "18888888888")
	if err != nil {
		t.Skipf("sign token: %v", err)
	}

	publicPaths := []string{
		"/rpg/quests",
		"/rpg/lottery/pool",
		"/rpg/pets/catalog",
		"/rpg/activities/current",
		"/rpg/guilds?page=1&pageSize=5",
		"/user/public/1",
		"/user/public/1/articles?page=1",
		"/rpg/public/1/status",
		"/rpg/leaderboard?type=exp&period=total&limit=5",
		"/rpg/weather-buff",
	}
	for _, path := range publicPaths {
		resp, status, err := c.GET(ctx, path, "")
		testutil.AssertOK(t, "public "+path, resp, status, err)
	}

	authedPaths := []string{
		"/rpg/status",
		"/rpg/sign-info",
		"/rpg/my-achievements",
		"/rpg/my-quests",
		"/rpg/my-buffs",
		"/rpg/ban-status",
		"/rpg/inventory",
		"/rpg/loadout",
		"/rpg/lottery/tickets",
		"/rpg/lottery/history",
	}
	for _, path := range authedPaths {
		resp, status, err := c.GET(ctx, path, token)
		testutil.AssertOK(t, "authed "+path, resp, status, err)
	}

	resp, status, err := c.GET(ctx, "/admin/rpg/stats", token)
	testutil.AssertOK(t, "admin rpg stats", resp, status, err)
}

func TestIntegrationAdminOps(t *testing.T) {
	c := newClient(t)
	ctx := context.Background()

	token, err := c.Login(ctx)
	token = testutil.SkipUnlessLogin(t, c, token, err)

	adminPaths := []string{
		"/admin/menu",
		"/role?page=1&pageSize=10",
		"/role/menu-privilege-tree",
		"/dept/tree",
		"/sensitive-word?page=1&pageSize=10",
		"/notification/list?page=1&pageSize=10",
		"/notification/unread-count",
		"/notification/since?seq=0",
		"/operation-log?page=1&pageSize=10",
	}
	for _, path := range adminPaths {
		resp, status, err := c.GET(ctx, path, token)
		testutil.AssertOK(t, path, resp, status, err)
	}

	var since []json.RawMessage
	resp, status, err := c.GET(ctx, "/notification/since?seq=0", token)
	testutil.AssertOK(t, "since array", resp, status, err)
	if err := testutil.UnmarshalData(resp, &since); err != nil {
		t.Fatal(err)
	}
}
