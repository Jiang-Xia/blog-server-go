// Package pub 公开统计接口；blog 域数据 Plan 05 前使用 mock。
package pub

import (
	"context"

	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/repo"
)

// Service 公开统计数据服务。
type Service struct {
	userRepo *repo.UserRepo
}

// NewService 构造 Service。
func NewService(userRepo *repo.UserRepo) *Service {
	return &Service{userRepo: userRepo}
}

// Stats 站点统计，与 Nest pub 契约对齐（blog 计数暂 mock）。
type Stats struct {
	ArticleCount  int `json:"articleCount"`
	CategoryCount int `json:"categoryCount"`
	TagCount      int `json:"tagCount"`
	UserCount     int `json:"userCount"`
}

// GetStats 返回 mock 文章/分类/标签数及真实用户数。
func (s *Service) GetStats(ctx context.Context) (*Stats, error) {
	userCount, err := s.userRepo.CountAll(ctx)
	if err != nil {
		return nil, err
	}
	return &Stats{
		ArticleCount:  128,
		CategoryCount: 12,
		TagCount:      36,
		UserCount:     userCount,
	}, nil
}
