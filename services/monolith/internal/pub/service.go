// Package pub 公开统计接口；聚合 blog 域计数与 user 总数。
package pub

import (
	"context"

	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent/article"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/repo"
)

// Service 公开统计数据服务。
type Service struct {
	ent      *ent.Client
	userRepo *repo.UserRepo
}

// NewService 构造 Service。
func NewService(client *ent.Client, userRepo *repo.UserRepo) *Service {
	return &Service{ent: client, userRepo: userRepo}
}

// Stats 站点统计，与 Nest pub 契约对齐。
type Stats struct {
	ArticleCount  int `json:"articleCount"`
	CategoryCount int `json:"categoryCount"`
	TagCount      int `json:"tagCount"`
	UserCount     int `json:"userCount"`
}

// GetStats 返回已发布文章/分类/标签数及用户总数。
func (s *Service) GetStats(ctx context.Context) (*Stats, error) {
	userCount, err := s.userRepo.CountAll(ctx)
	if err != nil {
		return nil, err
	}
	articleCount, err := s.ent.Article.Query().
		Where(article.IsDeleteEQ(false), article.StatusEQ("publish")).
		Count(ctx)
	if err != nil {
		return nil, err
	}
	categoryCount, err := s.ent.Category.Query().Count(ctx)
	if err != nil {
		return nil, err
	}
	tagCount, err := s.ent.Tag.Query().Count(ctx)
	if err != nil {
		return nil, err
	}
	return &Stats{
		ArticleCount:  articleCount,
		CategoryCount: categoryCount,
		TagCount:      tagCount,
		UserCount:     userCount,
	}, nil
}
