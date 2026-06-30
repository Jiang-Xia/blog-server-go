// Package operationlog 管理端操作日志记录与查询。
package operationlog

import (
	"context"
	"time"

	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent/operationlog"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/repo"
)

// CreateParams 写入操作日志参数。
type CreateParams struct {
	UserID      int
	Username    string
	Module      string
	Action      string
	Method      string
	Path        string
	Description string
	IP          string
	RequestBody string
	StatusCode  int
}

// Service 操作日志业务逻辑。
type Service struct {
	client *ent.Client
}

// NewService 构造操作日志服务。
func NewService(client *ent.Client) *Service {
	return &Service{client: client}
}

// Create 异步写入操作日志。
func (s *Service) Create(ctx context.Context, p CreateParams) error {
	b := s.client.OperationLog.Create().
		SetUserId(p.UserID).
		SetUsername(p.Username).
		SetModule(p.Module).
		SetAction(p.Action).
		SetMethod(p.Method).
		SetPath(p.Path).
		SetDescription(p.Description).
		SetIP(p.IP).
		SetStatusCode(p.StatusCode)
	if p.RequestBody != "" {
		b.SetRequestBody(p.RequestBody)
	}
	_, err := b.Save(ctx)
	return err
}

// ListQuery 操作日志列表查询参数。
type ListQuery struct {
	Page     int
	PageSize int
	Module   string
	Action   string
	Username string
	Keyword  string
}

// List 分页查询操作日志。
func (s *Service) List(ctx context.Context, q ListQuery) (map[string]interface{}, error) {
	page, pageSize := q.Page, q.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if page <= 0 {
		page = 1
	}
	query := s.client.OperationLog.Query()
	if q.Module != "" {
		query = query.Where(operationlog.ModuleEQ(q.Module))
	}
	if q.Action != "" {
		query = query.Where(operationlog.ActionEQ(q.Action))
	}
	if q.Username != "" {
		query = query.Where(operationlog.UsernameContains(q.Username))
	}
	if q.Keyword != "" {
		query = query.Where(
			operationlog.Or(
				operationlog.PathContains(q.Keyword),
				operationlog.DescriptionContains(q.Keyword),
			),
		)
	}
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, err
	}
	list, err := query.
		Order(ent.Desc(operationlog.FieldCreateTime)).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"list":       list,
		"pagination": repo.CalcNestPagination(total, pageSize, page),
	}, nil
}

// Clean 清理日志：days > 0 删除指定天数前，否则清空全部。
func (s *Service) Clean(ctx context.Context, days int) (int, error) {
	q := s.client.OperationLog.Delete()
	if days > 0 {
		cutoff := time.Now().AddDate(0, 0, -days)
		q = q.Where(operationlog.CreateTimeLT(cutoff))
	}
	n, err := q.Exec(ctx)
	if err != nil {
		return 0, err
	}
	return n, nil
}
