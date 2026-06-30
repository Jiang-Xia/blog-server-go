package service

import (
	"context"

	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
	blogrepo "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/blog/repo"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/blog/util"
	"github.com/google/uuid"
)

// TagService 标签 CRUD。
type TagService struct {
	tags     *blogrepo.TagRepo
	articles *blogrepo.ArticleRepo
}

// NewTagService 构造 TagService。
func NewTagService(tags *blogrepo.TagRepo, articles *blogrepo.ArticleRepo) *TagService {
	return &TagService{tags: tags, articles: articles}
}

// Create 创建标签。
func (s *TagService) Create(ctx context.Context, uid int, label, value string) (interface{}, error) {
	if label == "" {
		return nil, errcode.WithMessage(errcode.InvalidParam, "请输入标签名称")
	}
	exists, err := s.tags.ExistsByLabel(ctx, label)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errcode.WithMessage(errcode.InternalError, "标签已存在")
	}
	if value == "" {
		value = label
	}
	row, err := s.tags.Create(ctx, &ent.Tag{
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

// List 标签列表（含 articleCount）。
func (s *TagService) List(ctx context.Context, title string, countPublishedOnly bool) ([]map[string]interface{}, error) {
	rows, err := s.tags.List(ctx, title)
	if err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, 0, len(rows))
	for _, t := range rows {
		count, err := s.articles.CountByTag(ctx, t.ID, countPublishedOnly)
		if err != nil {
			return nil, err
		}
		out = append(out, tagToMap(t, count))
	}
	sortByArticleCount(out)
	return out, nil
}

// Get 标签详情。
func (s *TagService) Get(ctx context.Context, id string) (interface{}, error) {
	row, err := s.tags.FindByID(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errcode.WithMessage(errcode.NotFound, "标签不存在")
		}
		return nil, err
	}
	return row, nil
}

// GetWithArticles 标签详情含关联文章。
func (s *TagService) GetWithArticles(ctx context.Context, id, status string) (interface{}, error) {
	tagRow, err := s.tags.FindByID(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errcode.WithMessage(errcode.NotFound, "标签不存在")
		}
		return nil, err
	}
	articles, err := s.articles.ListByTagID(ctx, tagRow.ID, status)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"id":       tagRow.ID,
		"uid":      tagRow.UID,
		"label":    tagRow.Label,
		"value":    tagRow.Value,
		"color":    tagRow.Color,
		"createAt": tagRow.CreateAt,
		"updateAt": tagRow.UpdateAt,
		"articles": articles,
	}, nil
}

// Update 更新标签。
func (s *TagService) Update(ctx context.Context, id string, label, value, color *string) (interface{}, error) {
	if _, err := s.tags.FindByID(ctx, id); err != nil {
		if ent.IsNotFound(err) {
			return nil, errcode.WithMessage(errcode.NotFound, "标签不存在")
		}
		return nil, err
	}
	row, err := s.tags.Update(ctx, id, label, value, color)
	if err != nil {
		return nil, err
	}
	return row, nil
}

// Delete 删除标签。
func (s *TagService) Delete(ctx context.Context, id string) error {
	if _, err := s.tags.FindByID(ctx, id); err != nil {
		if ent.IsNotFound(err) {
			return errcode.WithMessage(errcode.NotFound, "标签不存在")
		}
		return err
	}
	if err := s.tags.Delete(ctx, id); err != nil {
		return errcode.WithMessage(errcode.InternalError, "删除失败，可能存在关联文章")
	}
	return nil
}

// FindByIDs 批量查询标签。
func (s *TagService) FindByIDs(ctx context.Context, ids []string) ([]*ent.Tag, error) {
	return s.tags.FindByIDs(ctx, ids)
}

func tagToMap(t *ent.Tag, count int) map[string]interface{} {
	return map[string]interface{}{
		"id":           t.ID,
		"uid":          t.UID,
		"label":        t.Label,
		"value":        t.Value,
		"color":        t.Color,
		"createAt":     t.CreateAt,
		"updateAt":     t.UpdateAt,
		"articleCount": count,
	}
}
