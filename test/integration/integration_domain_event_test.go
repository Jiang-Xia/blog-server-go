//go:build integration

package integration_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Jiang-Xia/blog-server-go/test/testutil"
)

// TestIntegrationCommentPublishesRPGExp 发评论后 RPG 经验应增加（依赖 blog:events → rpg-handlers）。
func TestIntegrationCommentPublishesRPGExp(t *testing.T) {
	c := newClient(t)
	ctx := context.Background()
	token := testutil.MustSignToken(t, 1, "18888888888")

	before, status, err := c.GET(ctx, "/rpg/status", token)
	testutil.AssertOK(t, "rpg status before", before, status, err)
	var beforeData struct {
		Exp int `json:"exp"`
	}
	if err := testutil.UnmarshalData(before, &beforeData); err != nil {
		t.Fatal(err)
	}

	list, status, err := c.POST(ctx, "/article/list", map[string]any{
		"page": 1, "pageSize": 1, "client": true,
	}, token)
	testutil.AssertOK(t, "article list", list, status, err)
	var articles struct {
		List []struct {
			ID int `json:"id"`
		} `json:"list"`
	}
	if err := testutil.UnmarshalData(list, &articles); err != nil || len(articles.List) == 0 {
		t.Skip("无文章数据，跳过领域事件联调")
	}
	articleID := articles.List[0].ID

	content := fmt.Sprintf("[integration-event] %d", time.Now().UnixNano())
	comment, status, err := c.POST(ctx, "/comment/create", map[string]any{
		"articleId": articleID, "content": content,
	}, token)
	testutil.AssertOK(t, "comment create", comment, status, err)

	// 等待 rpg-service Stream 消费（异步）。
	time.Sleep(500 * time.Millisecond)

	after, status, err := c.GET(ctx, "/rpg/status", token)
	testutil.AssertOK(t, "rpg status after", after, status, err)
	var afterData struct {
		Exp int `json:"exp"`
	}
	if err := testutil.UnmarshalData(after, &afterData); err != nil {
		t.Fatal(err)
	}
	if afterData.Exp < beforeData.Exp+5 {
		t.Fatalf("exp should increase by at least 5: before=%d after=%d", beforeData.Exp, afterData.Exp)
	}
}
