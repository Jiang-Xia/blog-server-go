// Package userport blog 对 user 域跨服务依赖（经 user Kitex）。
package userport

import (
	"context"
	"fmt"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/usersvc"
)

// GRPCArticleUserPort 经 user Kitex 获取用户摘要与部门信息。
type GRPCArticleUserPort struct {
	scope usersvc.ArticleScope
	users usersvc.UserService
}

// NewGRPCArticleUserPort 构造 ArticleUserPort。
func NewGRPCArticleUserPort(client usersvc.CrossClient) *GRPCArticleUserPort {
	return &GRPCArticleUserPort{scope: client, users: client}
}

// ListActiveUserIDs 返回 active 用户 ID 列表，C 端文章列表过滤锁定作者。
func (p *GRPCArticleUserPort) ListActiveUserIDs(ctx context.Context) ([]int, error) {
	return p.scope.ListActiveUserIDs(ctx)
}

// FindUserForArticle 获取作者信息（含 deptId）。
func (p *GRPCArticleUserPort) FindUserForArticle(ctx context.Context, uid int) (*ArticleAuthorUser, error) {
	u, err := p.users.GetUser(ctx, uint64(uid))
	if err != nil {
		return nil, err
	}
	return &ArticleAuthorUser{ID: uid, DeptID: u.DeptID, Status: u.Status}, nil
}

// FindDeptByID 机构名查询。
func (p *GRPCArticleUserPort) FindDeptByID(ctx context.Context, id int) (*DeptInfo, error) {
	d, err := p.scope.GetDept(ctx, id)
	if err != nil {
		return nil, err
	}
	return &DeptInfo{ID: d.ID, DeptName: d.DeptName}, nil
}

// GRPCArticleAdminPort 经 user Kitex 解析文章数据权限。
type GRPCArticleAdminPort struct {
	scope usersvc.ArticleScope
}

// NewGRPCArticleAdminPort 构造 ArticleAdminPort。
func NewGRPCArticleAdminPort(client usersvc.ArticleScope) *GRPCArticleAdminPort {
	return &GRPCArticleAdminPort{scope: client}
}

// ResolveArticleAccessibleDeptIDs nil 表示不限制机构（超管）。
func (p *GRPCArticleAdminPort) ResolveArticleAccessibleDeptIDs(ctx context.Context, uid int) ([]int, error) {
	return p.scope.ResolveArticleAccessibleDeptIDs(ctx, uid)
}

// AssertArticleDeptAccess 校验机构数据权限。
func (p *GRPCArticleAdminPort) AssertArticleDeptAccess(ctx context.Context, uid int, deptID *int) error {
	return p.scope.AssertArticleDeptAccess(ctx, uid, deptID)
}

// ProvideUserService 装配 blog-service 专用 user Kitex 客户端（etcd 发现）。
func ProvideUserService(cfg *config.Config) (usersvc.CrossClient, error) {
	endpoints := cfg.Registry.EtcdEndpointsOrEmpty()
	if len(endpoints) == 0 {
		return nil, fmt.Errorf("registry.etcd_endpoints required for blog-service")
	}
	return usersvc.NewKitexUserService(endpoints)
}
