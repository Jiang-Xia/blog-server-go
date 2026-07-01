// Package aggregator gateway BFF 聚合接口。
package aggregator

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/response"
	"github.com/cloudwego/hertz/pkg/app"
)

// StatsHandler GET /pub/stats BFF：合并 blog 统计与 user 用户数。
type StatsHandler struct {
	userURL string
	blogURL string
	client  *http.Client
}

// NewStatsHandler 构造 pub/stats 聚合 handler。
func NewStatsHandler(cfg *config.Config) *StatsHandler {
	return &StatsHandler{
		userURL: strings.TrimSuffix(cfg.Proxy.UserURL, "/"),
		blogURL: strings.TrimSuffix(cfg.Proxy.BlogURL, "/"),
		client:  &http.Client{Timeout: 5 * time.Second},
	}
}

type pubStats struct {
	ArticleCount  int `json:"articleCount"`
	CategoryCount int `json:"categoryCount"`
	TagCount      int `json:"tagCount"`
	UserCount     int `json:"userCount"`
}

type apiEnvelope struct {
	Data pubStats `json:"data"`
}

// Stats 聚合各服务统计；blog 暂 mock，userCount 读 user-service。
func (h *StatsHandler) Stats(ctx context.Context, c *app.RequestContext) {
	stats := pubStats{ArticleCount: 128, CategoryCount: 12, TagCount: 36}
	if h.userURL != "" {
		if count, err := h.fetchUserCount(ctx); err == nil {
			stats.UserCount = count
		}
	}
	response.Success(ctx, c, stats)
}

func (h *StatsHandler) fetchUserCount(ctx context.Context) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.userURL+"/api/v1/user/list", strings.NewReader(`{"page":1,"pageSize":1}`))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := h.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var env struct {
		Data struct {
			Total int `json:"total"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		return 0, err
	}
	if env.Data.Total > 0 {
		return env.Data.Total, nil
	}
	return 0, fmt.Errorf("empty user total")
}
