package service

import (
	"context"

	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/services/blog/ent"
	blogrepo "github.com/Jiang-Xia/blog-server-go/services/blog/internal/blog/repo"
	"github.com/Jiang-Xia/blog-server-go/pkg/pagination"
	"github.com/google/uuid"
)

// CollectService 收藏业务逻辑。
type CollectService struct {
	collects   *blogrepo.CollectRepo
	articles   *blogrepo.ArticleRepo
	categories *CategoryService
	tags       *TagService
}

// NewCollectService 构造 CollectService。
func NewCollectService(
	collects *blogrepo.CollectRepo,
	articles *blogrepo.ArticleRepo,
	categories *CategoryService,
	tags *TagService,
) *CollectService {
	return &CollectService{
		collects:   collects,
		articles:   articles,
		categories: categories,
		tags:       tags,
	}
}

// ToggleCollect 收藏/取消收藏。
func (s *CollectService) ToggleCollect(ctx context.Context, articleID, uid int) (map[string]interface{}, error) {
	existing, err := s.collects.FindByArticleAndUID(ctx, articleID, uid)
	if err == nil && existing != nil {
		if _, err := s.collects.DeleteByID(ctx, existing.ID); err != nil {
			return nil, err
		}
		return map[string]interface{}{"collected": false}, nil
	}
	if err != nil && !ent.IsNotFound(err) {
		return nil, err
	}
	_, err = s.collects.Create(ctx, &ent.Collect{
		ID:        uuid.NewString(),
		ArticleId: articleID,
		UID:       uid,
	})
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"collected": true}, nil
}

// CancelCollect 按收藏记录 id 取消。
func (s *CollectService) CancelCollect(ctx context.Context, id string) (map[string]interface{}, error) {
	n, err := s.collects.DeleteByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, errcode.WithMessage(errcode.NotFound, "收藏记录不存在")
	}
	return map[string]interface{}{"message": "取消收藏成功"}, nil
}

// GetCollectList 用户收藏列表。
func (s *CollectService) GetCollectList(ctx context.Context, uid, page, pageSize int) (map[string]interface{}, error) {
	rows, total, err := s.collects.ListByUID(ctx, uid, page, pageSize)
	if err != nil {
		return nil, err
	}
	list := make([]map[string]interface{}, 0, len(rows))
	for _, c := range rows {
		art, err := s.articles.GetByID(ctx, c.ArticleId)
		if err != nil || art.IsDelete {
			continue
		}
		item := map[string]interface{}{
			"id": c.ID, "articleId": c.ArticleId, "createTime": c.CreateTime,
			"article": map[string]interface{}{
				"id": art.ID, "title": art.Title, "description": art.Description,
				"status": art.Status, "isDelete": art.IsDelete, "topping": art.Topping,
				"views": art.Views, "likes": art.Likes, "createTime": art.CreateTime,
			},
		}
		if art.Articles != nil && *art.Articles != "" {
			if cat, err := s.categories.Get(ctx, *art.Articles); err == nil {
				if cRow, ok := cat.(*ent.Category); ok {
					item["article"].(map[string]interface{})["category"] = map[string]interface{}{
						"id": cRow.ID, "label": cRow.Label,
					}
				}
			}
		}
		tagIDs, _ := s.articles.ListTagIDsByArticle(ctx, art.ID)
		tagRows, _ := s.tags.FindByIDs(ctx, tagIDs)
		tagList := make([]map[string]interface{}, 0, len(tagRows))
		for _, t := range tagRows {
			tagList = append(tagList, map[string]interface{}{"label": t.Label})
		}
		item["article"].(map[string]interface{})["tags"] = tagList
		list = append(list, item)
	}
	return map[string]interface{}{
		"list":       list,
		"pagination": pagination.CalcNestPagination(total, pageSize, page),
	}, nil
}

// CheckCollected 是否已收藏。
func (s *CollectService) CheckCollected(ctx context.Context, articleID, uid int) (map[string]interface{}, error) {
	collected, err := s.collects.IsCollected(ctx, articleID, uid)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"collected": collected}, nil
}

// GetCollectCount 文章收藏数。
func (s *CollectService) GetCollectCount(ctx context.Context, articleID int) (map[string]interface{}, error) {
	count, err := s.collects.CountByArticle(ctx, articleID)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"count": count}, nil
}
