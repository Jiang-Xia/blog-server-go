// Package blogsvc blog 文章 RPG 字段跨服务 Kitex 客户端。
package blogsvc

import (
	"context"
	"fmt"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/kitexreg"
	blogv1 "github.com/Jiang-Xia/blog-server-go/proto/kitex/blog/v1"
	"github.com/Jiang-Xia/blog-server-go/proto/kitex/blog/v1/articleservice"
	"github.com/cloudwego/kitex/client"
)

// ArticleRPGFields 文章 RPG 快照，对齐 x_article RPG 列。
type ArticleRPGFields struct {
	ArticleID        int
	AuthorUID        int
	Title            string
	ArticleExp       int
	ArticleLevel     int
	ReputationGained int
	IsMasterpiece    int
	TipTotal         int
}

// ArticleRPGStore rpg-service 读写文章 RPG 字段（经 blog Kitex）。
type ArticleRPGStore interface {
	GetArticleRPGFields(ctx context.Context, articleID int) (*ArticleRPGFields, error)
	UpdateArticleRPGFields(ctx context.Context, articleID int, exp, level, repGained, isMasterpiece int) error
	AddArticleTipTotal(ctx context.Context, articleID, amount int) error
}

type kitexArticleRPGStore struct {
	client articleservice.Client
}

// NewKitexArticleRPGStore 经 etcd 发现 blog-service；endpoints 为空时返回 noop（方法均报错）。
func NewKitexArticleRPGStore(endpoints []string) (ArticleRPGStore, error) {
	if len(endpoints) == 0 {
		return noopArticleRPGStore{}, nil
	}
	r, err := kitexreg.NewResolver(endpoints)
	if err != nil {
		return nil, err
	}
	cli, err := articleservice.NewClient(config.KitexServiceBlog, client.WithResolver(r))
	if err != nil {
		return nil, fmt.Errorf("new blog kitex client: %w", err)
	}
	return &kitexArticleRPGStore{client: cli}, nil
}

func (g *kitexArticleRPGStore) GetArticleRPGFields(ctx context.Context, articleID int) (*ArticleRPGFields, error) {
	res, err := g.client.GetArticleRPGFields(ctx, &blogv1.GetArticleRPGFieldsRequest{ArticleId: int32(articleID)})
	if err != nil {
		return nil, err
	}
	return &ArticleRPGFields{
		ArticleID:        int(res.GetArticleId()),
		AuthorUID:        int(res.GetAuthorUid()),
		Title:            res.GetTitle(),
		ArticleExp:       int(res.GetArticleExp()),
		ArticleLevel:     int(res.GetArticleLevel()),
		ReputationGained: int(res.GetReputationGained()),
		IsMasterpiece:    int(res.GetIsMasterpiece()),
		TipTotal:         int(res.GetTipTotal()),
	}, nil
}

func (g *kitexArticleRPGStore) UpdateArticleRPGFields(ctx context.Context, articleID int, exp, level, repGained, isMasterpiece int) error {
	_, err := g.client.UpdateArticleRPGFields(ctx, &blogv1.UpdateArticleRPGFieldsRequest{
		ArticleId:        int32(articleID),
		ArticleExp:       int32(exp),
		ArticleLevel:     int32(level),
		ReputationGained: int32(repGained),
		IsMasterpiece:    int32(isMasterpiece),
	})
	return err
}

func (g *kitexArticleRPGStore) AddArticleTipTotal(ctx context.Context, articleID, amount int) error {
	_, err := g.client.AddArticleTipTotal(ctx, &blogv1.AddArticleTipTotalRequest{
		ArticleId: int32(articleID),
		Amount:    int32(amount),
	})
	return err
}

type noopArticleRPGStore struct{}

func (noopArticleRPGStore) GetArticleRPGFields(context.Context, int) (*ArticleRPGFields, error) {
	return nil, fmt.Errorf("blog kitex etcd endpoints not configured")
}
func (noopArticleRPGStore) UpdateArticleRPGFields(context.Context, int, int, int, int, int) error {
	return fmt.Errorf("blog kitex etcd endpoints not configured")
}
func (noopArticleRPGStore) AddArticleTipTotal(context.Context, int, int) error {
	return fmt.Errorf("blog kitex etcd endpoints not configured")
}
