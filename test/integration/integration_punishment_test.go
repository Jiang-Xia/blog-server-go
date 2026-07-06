//go:build integration

package integration_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Jiang-Xia/blog-server-go/test/testutil"
)

// TestIntegrationSensitiveWordPunishment 发含敏感词评论后 lifeValue 下降（需库中有 hpPenalty>0 的敏感词）。
func TestIntegrationSensitiveWordPunishment(t *testing.T) {
	c := newClient(t)
	ctx := context.Background()
	token := testutil.MustSignToken(t, 1, "18888888888")

	before, status, err := c.GET(ctx, "/rpg/status", token)
	testutil.AssertOK(t, "rpg status before", before, status, err)
	var beforeData struct {
		LifeValue int `json:"lifeValue"`
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
		t.Skip("无文章数据，跳过惩罚链联调")
	}
	articleID := articles.List[0].ID

	// 常见测试敏感词；若库未配置则跳过。
	content := fmt.Sprintf("integration敏感词测试 %d", time.Now().UnixNano())
	comment, status, err := c.POST(ctx, "/comment/create", map[string]any{
		"articleId": articleID, "content": content,
	}, token)
	if err != nil {
		t.Fatal(err)
	}
	if status == 403 && strings.Contains(string(comment), "禁言") {
		t.Skip("用户已被禁言，跳过")
	}
	if status != 200 && status != 201 {
		t.Skip("评论未成功（可能无敏感词配置），跳过惩罚链联调")
	}

	time.Sleep(800 * time.Millisecond)

	after, status, err := c.GET(ctx, "/rpg/status", token)
	testutil.AssertOK(t, "rpg status after", after, status, err)
	var afterData struct {
		LifeValue          int         `json:"lifeValue"`
		SensitiveHitsCount int         `json:"sensitiveHitsCount"`
		BanEndTime         interface{} `json:"banEndTime"`
	}
	if err := testutil.UnmarshalData(after, &afterData); err != nil {
		t.Fatal(err)
	}
	if afterData.LifeValue >= beforeData.LifeValue && afterData.SensitiveHitsCount == 0 {
		t.Skip("未触发敏感词扣 HP（库中无匹配词），跳过")
	}
	if afterData.LifeValue >= beforeData.LifeValue {
		t.Fatalf("lifeValue should decrease: before=%d after=%d hits=%d",
			beforeData.LifeValue, afterData.LifeValue, afterData.SensitiveHitsCount)
	}
}
