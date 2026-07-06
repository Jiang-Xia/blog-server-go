// Package blogsvc blog 文章 RPG 字段跨服务 gRPC 客户端。
package blogsvc

import (
	"context"
	"fmt"

	blogv1 "github.com/Jiang-Xia/blog-server-go/proto/gen/go/blog/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

// ArticleRPGStore rpg-service 读写文章 RPG 字段（经 blog gRPC）。
type ArticleRPGStore interface {
	GetArticleRPGFields(ctx context.Context, articleID int) (*ArticleRPGFields, error)
	UpdateArticleRPGFields(ctx context.Context, articleID int, exp, level, repGained, isMasterpiece int) error
	AddArticleTipTotal(ctx context.Context, articleID, amount int) error
}

type grpcArticleRPGStore struct {
	client blogv1.ArticleServiceClient
}

// NewGRPCArticleRPGStore 连接 blog-service gRPC。
func NewGRPCArticleRPGStore(addr string) (ArticleRPGStore, error) {
	if addr == "" {
		return noopArticleRPGStore{}, nil
	}
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("dial blog grpc %s: %w", addr, err)
	}
	return &grpcArticleRPGStore{client: blogv1.NewArticleServiceClient(conn)}, nil
}

func (g *grpcArticleRPGStore) GetArticleRPGFields(ctx context.Context, articleID int) (*ArticleRPGFields, error) {
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

func (g *grpcArticleRPGStore) UpdateArticleRPGFields(ctx context.Context, articleID int, exp, level, repGained, isMasterpiece int) error {
	_, err := g.client.UpdateArticleRPGFields(ctx, &blogv1.UpdateArticleRPGFieldsRequest{
		ArticleId:        int32(articleID),
		ArticleExp:       int32(exp),
		ArticleLevel:     int32(level),
		ReputationGained: int32(repGained),
		IsMasterpiece:    int32(isMasterpiece),
	})
	return err
}

func (g *grpcArticleRPGStore) AddArticleTipTotal(ctx context.Context, articleID, amount int) error {
	_, err := g.client.AddArticleTipTotal(ctx, &blogv1.AddArticleTipTotalRequest{
		ArticleId: int32(articleID),
		Amount:    int32(amount),
	})
	return err
}

type noopArticleRPGStore struct{}

func (noopArticleRPGStore) GetArticleRPGFields(context.Context, int) (*ArticleRPGFields, error) {
	return nil, fmt.Errorf("blog grpc addr not configured")
}
func (noopArticleRPGStore) UpdateArticleRPGFields(context.Context, int, int, int, int, int) error {
	return fmt.Errorf("blog grpc addr not configured")
}
func (noopArticleRPGStore) AddArticleTipTotal(context.Context, int, int) error {
	return fmt.Errorf("blog grpc addr not configured")
}
