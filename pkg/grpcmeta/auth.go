// Package grpcmeta 提供 gRPC 鉴权拦截器（内部服务从 metadata 取 userID，不重复验签）。
package grpcmeta

import (
	"context"

	"github.com/Jiang-Xia/blog-server-go/pkg/ctxutil"
	"google.golang.org/grpc"
)

// AuthUnaryInterceptor 从 metadata 注入 userID 到 context。
func AuthUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if uid := IncomingUserID(ctx); uid != 0 {
			ctx = ctxutil.WithUserID(ctx, int(uid))
		}
		return handler(ctx, req)
	}
}
