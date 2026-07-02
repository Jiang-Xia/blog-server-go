//go:build e2e

package e2e_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/Jiang-Xia/blog-server-go/test/testutil"
)

func TestE2EBlogLikeCollectAndBFF(t *testing.T) {
	c, err := testutil.NewClient("")
	if err != nil {
		t.Fatal(err)
	}
	testutil.RequireServer(t, c)
	ctx := context.Background()
	token := testutil.MustLogin(t, c)
	articleID := testutil.FirstArticleID(t, c, ctx, token)

	info, status, err := c.GET(ctx, fmt.Sprintf("/article/info?id=%d", articleID), "")
	testutil.AssertOK(t, "article/info BFF", info, status, err)

	like, status, err := c.POST(ctx, "/like", map[string]any{
		"articleId": articleID,
		"status":    true,
	}, token)
	testutil.AssertOK(t, "like add", like, status, err)

	check, status, err := c.GET(ctx, fmt.Sprintf("/like/check?articleId=%d", articleID), token)
	testutil.AssertOK(t, "like check after", check, status, err)

	collect, status, err := c.POST(ctx, "/collect", map[string]any{"articleId": articleID}, token)
	if err != nil || status != 200 {
		t.Fatalf("collect toggle: http=%d err=%v", status, err)
	}
	if !testutil.IsOK(collect) && collect.Code != 400 {
		t.Fatalf("collect code=%d msg=%s", collect.Code, collect.Message)
	}

	collect2, status, err := c.POST(ctx, "/collect", map[string]any{"articleId": articleID}, token)
	if err != nil || status != 200 {
		t.Fatalf("collect toggle again: http=%d err=%v", status, err)
	}
	if !testutil.IsOK(collect2) && collect2.Code != 400 {
		t.Fatalf("collect2 code=%d", collect2.Code)
	}

	views, status, err := c.POST(ctx, "/article/views", map[string]any{"articleId": articleID}, token)
	testutil.AssertOK(t, "article views", views, status, err)

	related, status, err := c.GET(ctx, fmt.Sprintf("/article/related?id=%d&limit=3", articleID), "")
	testutil.AssertOK(t, "article related", related, status, err)
}
