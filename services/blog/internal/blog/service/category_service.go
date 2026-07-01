// Package service 分类业务逻辑。
package service

import (
	"context"

	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/services/blog/ent"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/blog/domain"
	blogrepo "github.com/Jiang-Xia/blog-server-go/services/blog/internal/blog/repo"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/blog/util"
	"github.com/google/uuid"
)

// CategoryService 分类 CRUD。
type CategoryService struct {
	categories *blogrepo.CategoryRepo
	articles   *blogrepo.ArticleRepo
}

// NewCategoryService 构造 CategoryService。
func NewCategoryService(categories *blogrepo.CategoryRepo, articles *blogrepo.ArticleRepo) *CategoryService {
	return &CategoryService{categories: categories, articles: articles}
}

// Create 创建分类。
func (s *CategoryService) Create(ctx context.Context, uid int, label, value string) (interface{}, error) {
	if label == "" {
		return nil, errcode.WithMessage(errcode.InvalidParam, "请输入分类名称")
	}
	exists, err := s.categories.ExistsByLabel(ctx, label)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errcode.WithMessage(errcode.InternalError, "分类已存在")
	}
	if value == "" {
		value = label
	}
	row, err := s.categories.Create(ctx, &ent.Category{
		ID:    uuid.NewString(),
		UID:   uid,
		Label: label,
		Value: value,
		Color: util.RandomColor(),
	})
	if err != nil {
		return nil, err
	}
	return row, nil
}

// List 分类列表（含 articleCount）。
func (s *CategoryService) List(ctx context.Context, title string, countPublishedOnly bool) ([]map[string]interface{}, error) {
	rows, err := s.categories.List(ctx, title)
	if err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, 0, len(rows))
	for _, c := range rows {
		count, err := s.articles.CountByCategory(ctx, c.ID, countPublishedOnly)
		if err != nil {
			return nil, err
		}
		out = append(out, categoryToMap(c, count))
	}
	sortByArticleCount(out)
	return out, nil
}

// Get 分类详情。
func (s *CategoryService) Get(ctx context.Context, id string) (interface{}, error) {
	row, err := s.categories.FindByID(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errcode.WithMessage(errcode.NotFound, "分类不存在")
		}
		return nil, err
	}
	return row, nil
}

// Update 更新分类。
func (s *CategoryService) Update(ctx context.Context, id string, label, value, color *string) (interface{}, error) {
	if _, err := s.categories.FindByID(ctx, id); err != nil {
		if ent.IsNotFound(err) {
			return nil, errcode.WithMessage(errcode.NotFound, "分类不存在")
		}
		return nil, err
	}
	row, err := s.categories.Update(ctx, id, label, value, color)
	if err != nil {
		return nil, err
	}
	return row, nil
}

// Delete 删除分类。
func (s *CategoryService) Delete(ctx context.Context, id string) error {
	if _, err := s.categories.FindByID(ctx, id); err != nil {
		if ent.IsNotFound(err) {
			return errcode.WithMessage(errcode.NotFound, "分类不存在")
		}
		return err
	}
	if err := s.categories.Delete(ctx, id); err != nil {
		return errcode.WithMessage(errcode.InternalError, "删除失败，可能存在关联文章")
	}
	return nil
}

// FindByID 内部查询分类实体。
func (s *CategoryService) FindByID(ctx context.Context, id string) (*ent.Category, error) {
	return s.categories.FindByID(ctx, id)
}

// ToItem 转为 domain CategoryItem。
func (s *CategoryService) ToItem(c *ent.Category) *domain.CategoryItem {
	if c == nil {
		return nil
	}
	return &domain.CategoryItem{ID: c.ID, Label: c.Label, Value: c.Value, Color: c.Color}
}

func categoryToMap(c *ent.Category, count int) map[string]interface{} {
	return map[string]interface{}{
		"id":           c.ID,
		"uid":          c.UID,
		"label":        c.Label,
		"value":        c.Value,
		"color":        c.Color,
		"createAt":     c.CreateAt,
		"updateAt":     c.UpdateAt,
		"articleCount": count,
	}
}

func sortByArticleCount(items []map[string]interface{}) {
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			ci, _ := items[i]["articleCount"].(int)
			cj, _ := items[j]["articleCount"].(int)
			if cj > ci {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
}
