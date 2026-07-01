// Package aggregator gateway BFF 聚合接口。
package aggregator

import (
	"context"
	"encoding/json"

	blogv1 "github.com/Jiang-Xia/blog-server-go/proto/gen/go/blog/v1"
	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/pkg/response"
	"github.com/Jiang-Xia/blog-server-go/services/gateway/internal/grpcclient"
	"github.com/cloudwego/hertz/pkg/app"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ArticleHandler GET /article/info BFF：经 blog gRPC 取详情 JSON。
type ArticleHandler struct {
	clients *grpcclient.Clients
}

// NewArticleHandler 构造 article/info BFF handler。
func NewArticleHandler(clients *grpcclient.Clients) *ArticleHandler {
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
		response.Error(ctx, c, errcode.WithMessage(errcode.InternalError, "blog gRPC 未配置"))
		return
	}
	resp, err := h.clients.Blog.GetArticleDetail(ctx, &blogv1.GetArticleDetailRequest{Key: id})
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			response.Error(ctx, c, errcode.WithMessage(errcode.NotFound, "找不到文章"))
			return
		}
		response.Error(ctx, c, errcode.WithMessage(errcode.InternalError, err.Error()))
		return
	}
	var data interface{}
	if err := json.Unmarshal(resp.GetDetailJson(), &data); err != nil {
		response.Error(ctx, c, errcode.WithMessage(errcode.InternalError, "解析文章详情失败"))
		return
	}
	response.Success(ctx, c, data)
}
