// Package userport 封装 blog 对 user 域的跨服务依赖（gRPC / 后续 RPC 扩展）。
package userport

import "context"

// DeptInfo 机构摘要。
type DeptInfo struct {
	ID       int
	DeptName string
}

// ArticleAuthorUser 创建文章所需的用户字段（含机构）。
type ArticleAuthorUser struct {
	ID     int
	DeptID *int
	Status string
}

// ArticleUserPort 文章模块对用户域的只读依赖。
type ArticleUserPort interface {
	ListActiveUserIDs(ctx context.Context) ([]int, error)
	FindUserForArticle(ctx context.Context, uid int) (*ArticleAuthorUser, error)
	FindDeptByID(ctx context.Context, id int) (*DeptInfo, error)
}

// ArticleAdminPort 文章数据权限（机构范围）依赖。
type ArticleAdminPort interface {
	ResolveArticleAccessibleDeptIDs(ctx context.Context, uid int) ([]int, error)
	AssertArticleDeptAccess(ctx context.Context, uid int, deptID *int) error
}
