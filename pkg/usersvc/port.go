// Package usersvc 定义跨服务 UserService 只读端口与 DTO（blog/rpg 经 Kitex 消费）。
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
	DeptID   *int   `json:"deptId,omitempty"`
}

// DeptDTO 部门摘要。
type DeptDTO struct {
	ID       int    `json:"id"`
	DeptName string `json:"deptName"`
}

// FilterEvaluateResult 敏感词检测结果。
type FilterEvaluateResult struct {
	Content    string
	HitWords   []string
	HpPenalty  int
	NeedReview bool
	Rejected   bool
}

// FilterHitParams 敏感词命中记录参数。
type FilterHitParams struct {
	SourceType string
	SourceID   string
	Content    string
	HitWords   []string
	UID        *int
	IP         *string
}

// UserService 用户域只读端口。
type UserService interface {
	GetUser(ctx context.Context, id uint64) (*UserDTO, error)
	GetUserBatch(ctx context.Context, ids []uint64) ([]*UserDTO, error)
}

// ContentFilter 敏感词过滤（user-service Kitex）。
type ContentFilter interface {
	EvaluateContent(ctx context.Context, content string) (*FilterEvaluateResult, error)
	CreateHitRecord(ctx context.Context, params FilterHitParams) error
}

// ArticleScope 文章数据权限与部门查询。
type ArticleScope interface {
	ListActiveUserIDs(ctx context.Context) ([]int, error)
	GetDept(ctx context.Context, id int) (*DeptDTO, error)
	ResolveArticleAccessibleDeptIDs(ctx context.Context, uid int) ([]int, error)
	AssertArticleDeptAccess(ctx context.Context, uid int, deptID *int) error
}

// SensitiveHitLister RPG C 端敏感词命中分页。
type SensitiveHitLister interface {
	ListSensitiveWordHits(ctx context.Context, uid, page, pageSize int) (map[string]interface{}, error)
}

// CrossClient 跨服务 user Kitex 聚合端口。
type CrossClient interface {
	UserService
	ContentFilter
	ArticleScope
	SensitiveHitLister
}

// SystemEmailSender 系统邮件发送（user-service Kitex）。
type SystemEmailSender interface {
	SendSystemEmail(ctx context.Context, to, subject, htmlBody string) (bool, error)
}
