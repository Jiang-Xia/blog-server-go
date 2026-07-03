// Package grpcserver 实现 blog.v1.ArticleService gRPC 服务端。
package grpcserver

import (
	"context"
	"encoding/json"
	"strconv"

	blogv1 "github.com/Jiang-Xia/blog-server-go/proto/gen/go/blog/v1"
	"github.com/Jiang-Xia/blog-server-go/services/blog/ent"
	"github.com/Jiang-Xia/blog-server-go/services/blog/ent/article"
	blogsvc "github.com/Jiang-Xia/blog-server-go/services/blog/internal/blog/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Server 实现 ArticleService gRPC。
type Server struct {
	blogv1.UnimplementedArticleServiceServer
	articles   *blogsvc.ArticleService
	moderation *blogsvc.ModerationService
	ent        *ent.Client
}

// New 构造 gRPC ArticleService 实现。
func New(articles *blogsvc.ArticleService, moderation *blogsvc.ModerationService, client *ent.Client) *Server {
	return &Server{articles: articles, moderation: moderation, ent: client}
}

// GetArticle 按 ID 返回文章摘要。
func (s *Server) GetArticle(ctx context.Context, req *blogv1.GetArticleRequest) (*blogv1.GetArticleResponse, error) {
	row, err := s.articles.Info(ctx, strconv.FormatUint(req.GetId(), 10))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "article not found: %v", err)
	}
	var authorID uint64
	if row.Info.UserInfo != nil {
		authorID = uint64(row.Info.UserInfo.ID)
	}
	return &blogv1.GetArticleResponse{
		Id:       uint64(row.Info.ID),
		Title:    row.Info.Title,
		AuthorId: authorID,
	}, nil
}

// ListArticles 分页列表占位实现。
func (s *Server) ListArticles(ctx context.Context, req *blogv1.ListArticlesRequest) (*blogv1.ListArticlesResponse, error) {
	_ = ctx
	_ = req
	return &blogv1.ListArticlesResponse{}, nil
}

// GetArticleDetail 返回文章详情 JSON（与 HTTP /article/info 同构）。
func (s *Server) GetArticleDetail(ctx context.Context, req *blogv1.GetArticleDetailRequest) (*blogv1.GetArticleDetailResponse, error) {
	key := req.GetKey()
	if key == "" {
		return nil, status.Error(codes.InvalidArgument, "key required")
	}
	data, err := s.articles.Info(ctx, key)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "article not found: %v", err)
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "marshal detail: %v", err)
	}
	return &blogv1.GetArticleDetailResponse{DetailJson: raw}, nil
}

// GetPubStats 返回公开统计计数（gateway pub/stats BFF）。
func (s *Server) GetPubStats(ctx context.Context, _ *emptypb.Empty) (*blogv1.GetPubStatsResponse, error) {
	if s.ent == nil {
		return &blogv1.GetPubStatsResponse{}, nil
	}
	articleCount, err := s.ent.Article.Query().
		Where(article.IsDeleteEQ(false), article.StatusEQ("publish")).
		Count(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "count articles: %v", err)
	}
	categoryCount, err := s.ent.Category.Query().Count(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "count categories: %v", err)
	}
	tagCount, err := s.ent.Tag.Query().Count(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "count tags: %v", err)
	}
	return &blogv1.GetPubStatsResponse{
		ArticleCount:  int32(articleCount),
		CategoryCount: int32(categoryCount),
		TagCount:      int32(tagCount),
	}, nil
}

// UpdateContentModerationStatus 敏感词审核后同步来源实体状态。
func (s *Server) UpdateContentModerationStatus(ctx context.Context, req *blogv1.UpdateContentModerationStatusRequest) (*blogv1.UpdateContentModerationStatusResponse, error) {
	if s.moderation == nil {
		return &blogv1.UpdateContentModerationStatusResponse{}, nil
	}
	updated, err := s.moderation.UpdateContentModerationStatus(ctx, req.GetSourceType(), req.GetSourceId(), req.GetStatus())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "update moderation status: %v", err)
	}
	return &blogv1.UpdateContentModerationStatusResponse{Updated: updated}, nil
}
