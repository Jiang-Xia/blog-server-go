//go:build integration

package integration_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Jiang-Xia/blog-server-go/test/testutil"
)

// TestIntegrationArticleLevelComment 评论后作者文章 articleExp 应增加。
func TestIntegrationArticleLevelComment(t *testing.T) {
	c := newClient(t)
	ctx := context.Background()
	token := testutil.MustSignToken(t, 1, "18888888888")

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
		t.Skip("无文章数据，跳过文章等级联调")
	}
	articleID := articles.List[0].ID

	infoBefore, status, err := c.GET(ctx, fmt.Sprintf("/article/info?id=%d", articleID), token)
	testutil.AssertOK(t, "article info before", infoBefore, status, err)
	var before struct {
		Info struct {
			ArticleExp int `json:"articleExp"`
		} `json:"info"`
	}
	if err := testutil.UnmarshalData(infoBefore, &before); err != nil {
		t.Fatal(err)
	}

	content := fmt.Sprintf("[integration-article-level] %d", time.Now().UnixNano())
	comment, status, err := c.POST(ctx, "/comment/create", map[string]any{
		"articleId": articleID, "content": content,
	}, token)
	testutil.AssertOK(t, "comment create", comment, status, err)

	time.Sleep(800 * time.Millisecond)

	infoAfter, status, err := c.GET(ctx, fmt.Sprintf("/article/info?id=%d", articleID), token)
	testutil.AssertOK(t, "article info after", infoAfter, status, err)
	var after struct {
		Info struct {
			ArticleExp int `json:"articleExp"`
		} `json:"info"`
	}
	if err := testutil.UnmarshalData(infoAfter, &after); err != nil {
		t.Fatal(err)
	}
	if after.Info.ArticleExp < before.Info.ArticleExp+3 {
		t.Fatalf("author articleExp should increase by at least 3: before=%d after=%d",
			before.Info.ArticleExp, after.Info.ArticleExp)
	}
}
