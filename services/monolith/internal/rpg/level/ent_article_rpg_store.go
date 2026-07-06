// Package level 单体模式文章 RPG 字段 Ent 读写（不经 gRPC）。
package level

import (
	"context"

	"github.com/Jiang-Xia/blog-server-go/pkg/blogsvc"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
)

// EntArticleRPGStore 经 monolith Ent 读写 x_article RPG 列。
type EntArticleRPGStore struct {
	client *ent.Client
}

// NewEntArticleRPGStore 构造 EntArticleRPGStore。
func NewEntArticleRPGStore(client *ent.Client) blogsvc.ArticleRPGStore {
	return &EntArticleRPGStore{client: client}
}

func (s *EntArticleRPGStore) GetArticleRPGFields(ctx context.Context, articleID int) (*blogsvc.ArticleRPGFields, error) {
	row, err := s.client.Article.Get(ctx, articleID)
	if err != nil {
		return nil, err
	}
	return &blogsvc.ArticleRPGFields{
		ArticleID:        row.ID,
		AuthorUID:        row.UID,
		Title:            row.Title,
		ArticleExp:       row.ArticleExp,
		ArticleLevel:     row.ArticleLevel,
		ReputationGained: row.ReputationGained,
		IsMasterpiece:    row.IsMasterpiece,
		TipTotal:         row.TipTotal,
	}, nil
}

func (s *EntArticleRPGStore) UpdateArticleRPGFields(ctx context.Context, articleID int, exp, level, repGained, isMasterpiece int) error {
	_, err := s.client.Article.UpdateOneID(articleID).
		SetArticleExp(exp).
		SetArticleLevel(level).
		SetReputationGained(repGained).
		SetIsMasterpiece(isMasterpiece).
		Save(ctx)
	return err
}

func (s *EntArticleRPGStore) AddArticleTipTotal(ctx context.Context, articleID, amount int) error {
	_, err := s.client.Article.UpdateOneID(articleID).AddTipTotal(amount).Save(ctx)
	return err
}
