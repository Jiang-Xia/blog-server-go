//go:build e2e

package e2e_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Jiang-Xia/blog-server-go/test/testutil"
)

func TestE2ERPGSignFlow(t *testing.T) {
	c, err := testutil.NewClient("")
	if err != nil {
		t.Fatal(err)
	}
	testutil.RequireServer(t, c)
	ctx := context.Background()

	token, err := c.Login(ctx)
	token = testutil.SkipUnlessLogin(t, c, token, err)

	resp, status, err := c.GET(ctx, "/rpg/level-rewards", "")
	testutil.AssertOK(t, "level-rewards", resp, status, err)
	var rewards []map[string]any
	if err := testutil.UnmarshalData(resp, &rewards); err != nil || len(rewards) < 4 {
		t.Fatalf("level-rewards want >=4 items, got %d", len(rewards))
	}

	signInfo, status, err := c.GET(ctx, "/rpg/sign-info", token)
	testutil.AssertOK(t, "sign-info before", signInfo, status, err)
	var before struct {
		SignedToday bool `json:"signedToday"`
	}
	_ = testutil.UnmarshalData(signInfo, &before)

	if !before.SignedToday {
		sign, status, err := c.POST(ctx, "/rpg/sign", nil, token)
		testutil.AssertOK(t, "sign", sign, status, err)
		var signData struct {
			Success bool `json:"success"`
			Exp     int  `json:"exp"`
		}
		if err := testutil.UnmarshalData(sign, &signData); err != nil || !signData.Success || signData.Exp < 10 {
			t.Fatalf("sign result: %+v err=%v", signData, err)
		}
	}

	dup, status, err := c.POST(ctx, "/rpg/sign", nil, token)
	if err != nil || status != 200 || dup.Code != 400 {
		t.Fatalf("duplicate sign want code=400, got code=%d err=%v", dup.Code, err)
	}

	after, status, err := c.GET(ctx, "/rpg/sign-info", token)
	testutil.AssertOK(t, "sign-info after", after, status, err)
	var info struct {
		SignedToday   bool `json:"signedToday"`
		TotalSignDays int  `json:"totalSignDays"`
	}
	if err := testutil.UnmarshalData(after, &info); err != nil || !info.SignedToday || info.TotalSignDays < 1 {
		t.Fatalf("sign-info after: %+v", info)
	}

	st, status, err := c.GET(ctx, "/rpg/status", token)
	testutil.AssertOK(t, "status", st, status, err)
	var statusData struct {
		Exp       int  `json:"exp"`
		LifeValue int  `json:"lifeValue"`
		Level     int  `json:"level"`
		Frames    []any `json:"unlockedAvatarFrames"`
	}
	if err := testutil.UnmarshalData(st, &statusData); err != nil {
		t.Fatal(err)
	}
	if statusData.Exp < 10 || statusData.LifeValue != 100 || statusData.Level < 1 {
		t.Fatalf("rpg status: %+v", statusData)
	}

	noAuth, status, err := c.GET(ctx, "/rpg/status", "")
	if err != nil || !testutil.IsUnauthorized(noAuth) {
		t.Fatalf("no auth want 401, http=%d code=%d", status, noAuth.Code)
	}

	poster, status, err := c.POST(ctx, "/rpg/activities/share-poster", nil, token)
	testutil.AssertOK(t, "share-poster", poster, status, err)
	var posterData struct {
		ActivityCode string `json:"activityCode"`
	}
	if err := testutil.UnmarshalData(poster, &posterData); err != nil || posterData.ActivityCode == "" {
		t.Fatalf("share-poster missing activityCode")
	}
}

func TestE2EBlogInteractionFlow(t *testing.T) {
	c, err := testutil.NewClient("")
	if err != nil {
		t.Fatal(err)
	}
	testutil.RequireServer(t, c)
	ctx := context.Background()

	unauth, status, err := c.POST(ctx, "/comment/create", map[string]any{
		"articleId": 1, "content": "e2e-unauth",
	}, "")
	if err != nil || !testutil.IsUnauthorized(unauth) {
		t.Fatalf("unauth comment want 401, http=%d code=%d", status, unauth.Code)
	}

	token, err := c.Login(ctx)
	token = testutil.SkipUnlessLogin(t, c, token, err)

	list, status, err := c.POST(ctx, "/article/list", map[string]any{
		"page": 1, "pageSize": 5, "client": true,
	}, token)
	testutil.AssertOK(t, "article list", list, status, err)
	var articles struct {
		List []struct {
			ID int `json:"id"`
		} `json:"list"`
	}
	if err := testutil.UnmarshalData(list, &articles); err != nil || len(articles.List) == 0 {
		t.Skip("无文章数据，请先 sync-data")
	}
	articleID := articles.List[0].ID

	for _, path := range []string{
		fmt.Sprintf("/like/check?articleId=%d", articleID),
		fmt.Sprintf("/collect/count?articleId=%d", articleID),
		fmt.Sprintf("/comment/findAll?articleId=%d&page=1&pageSize=5", articleID),
	} {
		resp, status, err := c.GET(ctx, path, token)
		testutil.AssertOK(t, path, resp, status, err)
	}

	content := fmt.Sprintf("[E2E-GO] comment %d", time.Now().UnixNano())
	comment, status, err := c.POST(ctx, "/comment/create", map[string]any{
		"articleId": articleID,
		"content":   content,
	}, token)
	if err != nil {
		t.Fatal(err)
	}
	// 敏感词/频控等业务拒绝仍视为链路可达
	if !testutil.IsOK(comment) && comment.Code == 0 {
		t.Fatalf("comment unexpected: code=%d msg=%s", comment.Code, comment.Message)
	}

	since, status, err := c.GET(ctx, "/notification/since?seq=0", token)
	testutil.AssertOK(t, "notification since", since, status, err)
}
