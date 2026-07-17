// Package kitexserver 实现 blog.v1.ArticleService Kitex 服务端。
// 数据来源：ArticleService / ModerationService / Ent / publicprofile。
package kitexserver

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/Jiang-Xia/blog-server-go/pkg/publicprofile"
	blogv1 "github.com/Jiang-Xia/blog-server-go/proto/kitex/blog/v1"
	"github.com/Jiang-Xia/blog-server-go/services/blog/ent"
	"github.com/Jiang-Xia/blog-server-go/services/blog/ent/article"
	blogsvc "github.com/Jiang-Xia/blog-server-go/services/blog/internal/blog/service"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Server 实现 ArticleService Kitex 接口。
type Server struct {
	articles      *blogsvc.ArticleService
	moderation    *blogsvc.ModerationService
	ent           *ent.Client
	publicProfile *publicprofile.Repo
}

// New 构造 Kitex ArticleService 实现。
func New(articles *blogsvc.ArticleService, moderation *blogsvc.ModerationService, client *ent.Client, publicProfile *publicprofile.Repo) *Server {
	return &Server{articles: articles, moderation: moderation, ent: client, publicProfile: publicProfile}
}

// GetArticle 按 ID 返回文章摘要。
func (s *Server) GetArticle(ctx context.Context, req *blogv1.GetArticleRequest) (*blogv1.GetArticleResponse, error) {
	row, err := s.articles.Info(ctx, strconv.FormatUint(req.GetId(), 10))
	if err != nil {
		return nil, fmt.Errorf("not found: article not found: %w", err)
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
		return nil, fmt.Errorf("invalid argument: key required")
	}
	data, err := s.articles.Info(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("not found: article not found: %w", err)
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal detail: %w", err)
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
		return nil, fmt.Errorf("count articles: %w", err)
	}
	categoryCount, err := s.ent.Category.Query().Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("count categories: %w", err)
	}
	tagCount, err := s.ent.Tag.Query().Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("count tags: %w", err)
	}
	// 多实例学习：对照 docker logs 观察 Kitex 负载均衡。
	host, _ := os.Hostname()
	log.Printf("[kitex] GetPubStats instance=%s articles=%d", host, articleCount)
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
		return nil, fmt.Errorf("update moderation status: %w", err)
	}
	return &blogv1.UpdateContentModerationStatusResponse{Updated: updated}, nil
}

// GetArticleRPGFields 读取文章 RPG 字段。
func (s *Server) GetArticleRPGFields(ctx context.Context, req *blogv1.GetArticleRPGFieldsRequest) (*blogv1.GetArticleRPGFieldsResponse, error) {
	if s.ent == nil {
		return nil, fmt.Errorf("unavailable: ent client not configured")
	}
	id := int(req.GetArticleId())
	if id <= 0 {
		return nil, fmt.Errorf("invalid argument: article_id required")
	}
	row, err := s.ent.Article.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("not found: article not found")
		}
		return nil, fmt.Errorf("get article: %w", err)
	}
	return &blogv1.GetArticleRPGFieldsResponse{
		ArticleId:        int32(row.ID),
		AuthorUid:        int32(row.UID),
		Title:            row.Title,
		ArticleExp:       int32(row.ArticleExp),
		ArticleLevel:     int32(row.ArticleLevel),
		ReputationGained: int32(row.ReputationGained),
		IsMasterpiece:    int32(row.IsMasterpiece),
		TipTotal:         int32(row.TipTotal),
	}, nil
}

// UpdateArticleRPGFields 更新文章 RPG 字段。
func (s *Server) UpdateArticleRPGFields(ctx context.Context, req *blogv1.UpdateArticleRPGFieldsRequest) (*blogv1.UpdateArticleRPGFieldsResponse, error) {
	if s.ent == nil {
		return nil, fmt.Errorf("unavailable: ent client not configured")
	}
	id := int(req.GetArticleId())
	if id <= 0 {
		return nil, fmt.Errorf("invalid argument: article_id required")
	}
	n, err := s.ent.Article.UpdateOneID(id).
		SetArticleExp(int(req.GetArticleExp())).
		SetArticleLevel(int(req.GetArticleLevel())).
		SetReputationGained(int(req.GetReputationGained())).
		SetIsMasterpiece(int(req.GetIsMasterpiece())).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return &blogv1.UpdateArticleRPGFieldsResponse{Updated: false}, nil
		}
		return nil, fmt.Errorf("update article rpg: %w", err)
	}
	_ = n
	return &blogv1.UpdateArticleRPGFieldsResponse{Updated: true}, nil
}

// AddArticleTipTotal 原子累加打赏总额。
func (s *Server) AddArticleTipTotal(ctx context.Context, req *blogv1.AddArticleTipTotalRequest) (*blogv1.AddArticleTipTotalResponse, error) {
	if s.ent == nil {
		return nil, fmt.Errorf("unavailable: ent client not configured")
	}
	id := int(req.GetArticleId())
	amount := int(req.GetAmount())
	if id <= 0 || amount <= 0 {
		return nil, fmt.Errorf("invalid argument: article_id and amount required")
	}
	n, err := s.ent.Article.UpdateOneID(id).AddTipTotal(amount).Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return &blogv1.AddArticleTipTotalResponse{Updated: false}, nil
		}
		return nil, fmt.Errorf("add tip total: %w", err)
	}
	_ = n
	return &blogv1.AddArticleTipTotalResponse{Updated: true}, nil
}

// ListPublicCollectArticles 用户公开收藏文章分页。
func (s *Server) ListPublicCollectArticles(ctx context.Context, req *blogv1.ListPublicProfileArticlesRequest) (*blogv1.ListPublicProfileArticlesResponse, error) {
	if s.publicProfile == nil {
		return &blogv1.ListPublicProfileArticlesResponse{}, nil
	}
	uid := int(req.GetUid())
	page := int(req.GetPage())
	pageSize := int(req.GetPageSize())
	rows, total, err := s.publicProfile.ListCollectArticles(ctx, uid, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("list public collects: %w", err)
	}
	return rowsToProto(rows, total, page, pageSize), nil
}

// ListPublicLikeArticles 用户公开点赞文章分页。
func (s *Server) ListPublicLikeArticles(ctx context.Context, req *blogv1.ListPublicProfileArticlesRequest) (*blogv1.ListPublicProfileArticlesResponse, error) {
	if s.publicProfile == nil {
		return &blogv1.ListPublicProfileArticlesResponse{}, nil
	}
	uid := int(req.GetUid())
	page := int(req.GetPage())
	pageSize := int(req.GetPageSize())
	rows, total, err := s.publicProfile.ListLikeArticles(ctx, uid, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("list public likes: %w", err)
	}
	return rowsToProto(rows, total, page, pageSize), nil
}

func rowsToProto(rows []publicprofile.ArticleRow, total, page, pageSize int) *blogv1.ListPublicProfileArticlesResponse {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	items := make([]*blogv1.PublicProfileArticleItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, &blogv1.PublicProfileArticleItem{
			Id:            int32(row.ID),
			Title:         row.Title,
			Description:   row.Description,
			Cover:         row.Cover,
			Views:         int32(row.Views),
			Likes:         int32(row.Likes),
			ArticleLevel:  int32(row.ArticleLevel),
			IsMasterpiece: int32(row.IsMasterpiece),
			TipTotal:      int32(row.TipTotal),
			CreateTime:    row.CreateTime.UTC().Format(time.RFC3339Nano),
		})
	}
	return &blogv1.ListPublicProfileArticlesResponse{
		Items:    items,
		Total:    int32(total),
		Page:     int32(page),
		PageSize: int32(pageSize),
	}
}
