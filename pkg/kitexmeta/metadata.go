// Package kitexmeta 定义微服务间 Kitex metainfo 键与读写辅助（替代原 gRPC metadata）。
package kitexmeta

import (
	"context"
	"strconv"

	"github.com/bytedance/gopkg/cloud/metainfo"
)

const (
	// UserIDKey gateway 验签后透传的用户 ID。
	UserIDKey = "x-user-id"
)

// OutgoingUserID 将 userID 写入 outgoing metainfo（客户端调用前）。
func OutgoingUserID(ctx context.Context, userID uint64) context.Context {
	if userID == 0 {
		return ctx
	}
	return metainfo.WithValue(ctx, UserIDKey, strconv.FormatUint(userID, 10))
}

// IncomingUserID 从 metainfo 读取 userID（服务端）。
func IncomingUserID(ctx context.Context) uint64 {
	val, ok := metainfo.GetValue(ctx, UserIDKey)
	if !ok || val == "" {
		return 0
	}
	id, _ := strconv.ParseUint(val, 10, 64)
	return id
}
