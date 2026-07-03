// Package rpgsvc 提供 RPG 域跨服务 gRPC 客户端（BanGuard 等）。
package rpgsvc

import "context"

// BanChecker 禁言判定端口（blog-service BanGuard 调用）。
type BanChecker interface {
	AssertNotBanned(ctx context.Context, uid int) error
}
