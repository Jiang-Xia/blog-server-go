// Package grpcmeta 定义微服务间 gRPC metadata 键与读写辅助。
package grpcmeta

import (
	"context"
	"strconv"

	"google.golang.org/grpc/metadata"
)

const (
	// UserIDKey gateway 验签后透传的用户 ID。
	UserIDKey = "x-user-id"
)

// OutgoingUserID 将 userID 写入 outgoing metadata。
func OutgoingUserID(ctx context.Context, userID uint64) context.Context {
	if userID == 0 {
		return ctx
	}
	return metadata.AppendToOutgoingContext(ctx, UserIDKey, strconv.FormatUint(userID, 10))
}

// IncomingUserID 从 incoming metadata 读取 userID。
func IncomingUserID(ctx context.Context) uint64 {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0
	}
	vals := md.Get(UserIDKey)
	if len(vals) == 0 {
		return 0
	}
	id, _ := strconv.ParseUint(vals[0], 10, 64)
	return id
}
