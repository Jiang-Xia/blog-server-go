// Package aggregator gateway BFF 聚合接口。
package aggregator

import (
	"context"
	"log"
	"os"

	"github.com/Jiang-Xia/blog-server-go/pkg/response"
	"github.com/Jiang-Xia/blog-server-go/services/gateway/internal/kitexclient"
	"github.com/cloudwego/hertz/pkg/app"
	"google.golang.org/protobuf/types/known/emptypb"
)

// StatsHandler GET /pub/stats BFF：合并 blog 统计与 user 用户数。
type StatsHandler struct {
	clients *kitexclient.Clients
}

// NewStatsHandler 构造 pub/stats 聚合 handler。
func NewStatsHandler(clients *kitexclient.Clients) *StatsHandler {
	return &StatsHandler{clients: clients}
}

type pubStats struct {
	ArticleCount  int `json:"articleCount"`
	CategoryCount int `json:"categoryCount"`
	TagCount      int `json:"tagCount"`
	UserCount     int `json:"userCount"`
}

// Stats 经 Kitex 聚合各服务统计。
func (h *StatsHandler) Stats(ctx context.Context, c *app.RequestContext) {
	// 多实例学习：对照 docker logs 观察 edge → gateway 负载。
	host, _ := os.Hostname()
	log.Printf("[bff] pub/stats gateway_instance=%s", host)
	stats := pubStats{}
	if h.clients != nil && h.clients.Blog != nil {
		if blogStats, err := h.clients.Blog.GetPubStats(ctx, &emptypb.Empty{}); err == nil && blogStats != nil {
			stats.ArticleCount = int(blogStats.GetArticleCount())
			stats.CategoryCount = int(blogStats.GetCategoryCount())
			stats.TagCount = int(blogStats.GetTagCount())
		}
	}
	if h.clients != nil && h.clients.User != nil {
		if userStats, err := h.clients.User.CountUsers(ctx, &emptypb.Empty{}); err == nil && userStats != nil {
			stats.UserCount = int(userStats.GetTotal())
		}
	}
	response.Success(ctx, c, stats)
}
