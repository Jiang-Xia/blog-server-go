package kitexmeta

import (
	"context"

	"github.com/Jiang-Xia/blog-server-go/pkg/ctxutil"
	"github.com/cloudwego/kitex/pkg/endpoint"
)

// AuthMiddleware 从 metainfo 注入 userID 到 context（内部服务不重复验签）。
func AuthMiddleware(next endpoint.Endpoint) endpoint.Endpoint {
	return func(ctx context.Context, req, resp any) (err error) {
		if uid := IncomingUserID(ctx); uid != 0 {
			ctx = ctxutil.WithUserID(ctx, int(uid))
		}
		return next(ctx, req, resp)
	}
}
