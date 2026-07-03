// Package usersvc 定义跨服务 UserService 只读端口与 DTO（blog/rpg 经 gRPC 消费）。
package usersvc

import "context"

// UserDTO 供 article 等模块消费的用户摘要。
type UserDTO struct {
	ID       uint64 `json:"id"`
	Nickname string `json:"nickname"`
	Username string `json:"username,omitempty"`
	Avatar   string `json:"avatar"`
	Email    string `json:"email,omitempty"`
	Status   string `json:"status,omitempty"`
}

// UserService 用户域只读端口。
type UserService interface {
	GetUser(ctx context.Context, id uint64) (*UserDTO, error)
	GetUserBatch(ctx context.Context, ids []uint64) ([]*UserDTO, error)
}

// SystemEmailSender 系统邮件发送（user-service gRPC）。
type SystemEmailSender interface {
	SendSystemEmail(ctx context.Context, to, subject, htmlBody string) (bool, error)
}
