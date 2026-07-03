// Package userport RPG 对用户域的跨服务只读依赖（经 user gRPC）。
package userport

import (
	"context"
	"fmt"
	"time"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/usersvc"
)

// UserInfo 公开主页/排行榜所需用户摘要。
type UserInfo struct {
	ID         int
	Nickname   string
	Username   *string
	Avatar     string
	Intro      string
	Status     string
	IsDelete   bool
	CreateTime time.Time
}

// UserReader 用户只读端口。
type UserReader interface {
	FindByID(ctx context.Context, uid int) (*UserInfo, error)
}

// GRPCUserReader 经 pkg/usersvc 实现 UserReader。
type GRPCUserReader struct {
	users usersvc.UserService
}

// NewGRPCUserReader 构造 UserReader。
func NewGRPCUserReader(users usersvc.UserService) *GRPCUserReader {
	return &GRPCUserReader{users: users}
}

// FindByID 按 ID 查询用户摘要。
func (r *GRPCUserReader) FindByID(ctx context.Context, uid int) (*UserInfo, error) {
	u, err := r.users.GetUser(ctx, uint64(uid))
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, fmt.Errorf("user %d not found", uid)
	}
	var username *string
	if u.Username != "" {
		s := u.Username
		username = &s
	}
	return &UserInfo{
		ID:       uid,
		Nickname: u.Nickname,
		Username: username,
		Avatar:   u.Avatar,
		Status:   u.Status,
	}, nil
}

// ProvideUserService 装配 rpg-service user gRPC 客户端。
func ProvideUserService(cfg *config.Config) (usersvc.CrossClient, error) {
	addr := cfg.GRPC.UserAddr
	if addr == "" {
		return nil, fmt.Errorf("GRPC.UserAddr required for rpg-service")
	}
	return usersvc.NewGRPCUserService(addr)
}
