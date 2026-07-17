// Package blogsvc 提供 blog 域跨服务 Kitex 客户端（敏感词审核联动等）。
package blogsvc

import "context"

// ContentModerationSyncer 审核后同步 comment/msgboard/reply 状态。
type ContentModerationSyncer interface {
	UpdateContentModerationStatus(ctx context.Context, sourceType, sourceID, status string) error
}
