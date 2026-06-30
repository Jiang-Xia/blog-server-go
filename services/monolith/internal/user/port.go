// Package user 定义跨模块 UserService 端口与本地实现入口。
package user

import "context"

// UserDTO 供 article 等模块消费的用户摘要（Plan 10 gRPC 前本地 interface）。
type UserDTO struct {
	ID       uint64 `json:"id"`
	Nickname string `json:"nickname"`
	Username string `json:"username,omitempty"`
	Avatar   string `json:"avatar"`
	Email    string `json:"email,omitempty"`
	Status   string `json:"status,omitempty"`
}

// UserService 用户域只读端口，Plan 05 起 article 等模块依赖此 interface。
type UserService interface {
	GetUser(ctx context.Context, id uint64) (*UserDTO, error)
	GetUserBatch(ctx context.Context, ids []uint64) ([]*UserDTO, error)
}
