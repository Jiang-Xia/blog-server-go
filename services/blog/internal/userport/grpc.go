package userport

import (
	"context"
	"fmt"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/usersvc"
)

// GRPCArticleUserPort 经 user gRPC 获取用户摘要（deptId 待 proto 扩展）。
type GRPCArticleUserPort struct {
	users usersvc.UserService
}

// NewGRPCArticleUserPort 构造 ArticleUserPort。
func NewGRPCArticleUserPort(users usersvc.UserService) *GRPCArticleUserPort {
	return &GRPCArticleUserPort{users: users}
}

// ListActiveUserIDs C 端列表过滤；gRPC 未扩展前返回 nil 表示不过滤。
func (p *GRPCArticleUserPort) ListActiveUserIDs(_ context.Context) ([]int, error) {
	return nil, nil
}

// FindUserForArticle 获取作者信息（deptId 暂不可用，创建文章会提示未关联机构）。
func (p *GRPCArticleUserPort) FindUserForArticle(ctx context.Context, uid int) (*ArticleAuthorUser, error) {
	u, err := p.users.GetUser(ctx, uint64(uid))
	if err != nil {
		return nil, err
	}
	return &ArticleAuthorUser{ID: uid, Status: u.Status}, nil
}

// FindDeptByID 机构名查询；待 user gRPC 扩展 dept RPC。
func (p *GRPCArticleUserPort) FindDeptByID(_ context.Context, _ int) (*DeptInfo, error) {
	return nil, fmt.Errorf("dept lookup not available via user gRPC yet")
}

// PermissiveArticleAdminPort 数据权限宽松实现（超管视角；待 admin gRPC）。
type PermissiveArticleAdminPort struct{}

// NewPermissiveArticleAdminPort 构造 ArticleAdminPort。
func NewPermissiveArticleAdminPort() *PermissiveArticleAdminPort {
	return &PermissiveArticleAdminPort{}
}

// ResolveArticleAccessibleDeptIDs nil 表示不限制机构。
func (p *PermissiveArticleAdminPort) ResolveArticleAccessibleDeptIDs(_ context.Context, _ int) ([]int, error) {
	return nil, nil
}

// AssertArticleDeptAccess 暂不校验机构权限。
func (p *PermissiveArticleAdminPort) AssertArticleDeptAccess(_ context.Context, _ int, _ *int) error {
	return nil
}

// ProvideUserService 装配 blog-service 专用 UserService gRPC 客户端。
func ProvideUserService(cfg *config.Config) (usersvc.UserService, error) {
	addr := cfg.GRPC.UserAddr
	if addr == "" {
		return nil, fmt.Errorf("GRPC.UserAddr required for blog-service")
	}
	return usersvc.NewGRPCUserService(addr)
}
