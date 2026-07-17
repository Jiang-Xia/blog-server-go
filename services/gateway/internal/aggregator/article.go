// Package aggregator gateway BFF 聚合接口。
package aggregator

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/pkg/response"
	blogv1 "github.com/Jiang-Xia/blog-server-go/proto/kitex/blog/v1"
	"github.com/Jiang-Xia/blog-server-go/services/gateway/internal/kitexclient"
	"github.com/cloudwego/hertz/pkg/app"
)

// ArticleHandler GET /article/info BFF：经 blog Kitex 取详情 JSON。
type ArticleHandler struct {
	clients *kitexclient.Clients
}

// NewArticleHandler 构造 article/info BFF handler。
func NewArticleHandler(clients *kitexclient.Clients) *ArticleHandler {
	return &ArticleHandler{clients: clients}
}

// Info 返回文章详情（与 blog-service HTTP 同构）。
func (h *ArticleHandler) Info(ctx context.Context, c *app.RequestContext) {
	id := string(c.Query("id"))
	if id == "" {
		response.Error(ctx, c, errcode.WithMessage(errcode.InvalidParam, "请输入有效 id"))
		return
	}
	if h.clients == nil || h.clients.Blog == nil {
		response.Error(ctx, c, errcode.WithMessage(errcode.InternalError, "blog Kitex 未配置"))
		return
	}
	resp, err := h.clients.Blog.GetArticleDetail(ctx, &blogv1.GetArticleDetailRequest{Key: id})
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			response.Error(ctx, c, errcode.WithMessage(errcode.NotFound, "找不到文章"))
			return
		}
		response.Error(ctx, c, errcode.WithMessage(errcode.InternalError, "%s", err.Error()))
		return
	}
	var data interface{}
	if err := json.Unmarshal(resp.GetDetailJson(), &data); err != nil {
		response.Error(ctx, c, errcode.WithMessage(errcode.InternalError, "解析文章详情失败"))
		return
	}
	response.Success(ctx, c, data)
}
