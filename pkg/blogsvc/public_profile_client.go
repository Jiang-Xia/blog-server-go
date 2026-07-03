package blogsvc

import (
	"context"
	"fmt"

	blogv1 "github.com/Jiang-Xia/blog-server-go/proto/gen/go/blog/v1"
	"github.com/Jiang-Xia/blog-server-go/pkg/pagination"
	"github.com/Jiang-Xia/blog-server-go/pkg/publicprofile"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// PublicProfileLister 公开主页收藏/点赞列表（blog gRPC）。
type PublicProfileLister interface {
	ListPublicCollectArticles(ctx context.Context, uid, page, pageSize int) (publicprofile.ListResult, error)
	ListPublicLikeArticles(ctx context.Context, uid, page, pageSize int) (publicprofile.ListResult, error)
}

type grpcPublicProfileLister struct {
	client blogv1.ArticleServiceClient
}

// NewGRPCPublicProfileLister 连接 blog-service gRPC。
func NewGRPCPublicProfileLister(addr string) (PublicProfileLister, error) {
	if addr == "" {
		return nil, fmt.Errorf("GRPC.BlogAddr required for rpg public profile lists")
	}
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("dial blog grpc %s: %w", addr, err)
	}
	return &grpcPublicProfileLister{client: blogv1.NewArticleServiceClient(conn)}, nil
}

func (g *grpcPublicProfileLister) ListPublicCollectArticles(ctx context.Context, uid, page, pageSize int) (publicprofile.ListResult, error) {
	resp, err := g.client.ListPublicCollectArticles(ctx, &blogv1.ListPublicProfileArticlesRequest{
		Uid: uint64(uid), Page: int32(page), PageSize: int32(pageSize),
	})
	if err != nil {
		return publicprofile.ListResult{}, err
	}
	return protoToListResult(resp), nil
}

func (g *grpcPublicProfileLister) ListPublicLikeArticles(ctx context.Context, uid, page, pageSize int) (publicprofile.ListResult, error) {
	resp, err := g.client.ListPublicLikeArticles(ctx, &blogv1.ListPublicProfileArticlesRequest{
		Uid: uint64(uid), Page: int32(page), PageSize: int32(pageSize),
	})
	if err != nil {
		return publicprofile.ListResult{}, err
	}
	return protoToListResult(resp), nil
}

func protoToListResult(resp *blogv1.ListPublicProfileArticlesResponse) publicprofile.ListResult {
	if resp == nil {
		return publicprofile.ListResult{List: []map[string]interface{}{}}
	}
	list := make([]map[string]interface{}, 0, len(resp.GetItems()))
	for _, item := range resp.GetItems() {
		list = append(list, map[string]interface{}{
			"id":             int(item.GetId()),
			"title":          item.GetTitle(),
			"description":    item.GetDescription(),
			"cover":          item.GetCover(),
			"views":          int(item.GetViews()),
			"likes":          int(item.GetLikes()),
			"articleLevel":   int(item.GetArticleLevel()),
			"isMasterpiece":  int(item.GetIsMasterpiece()),
			"tipTotal":       int(item.GetTipTotal()),
			"createTime":     item.GetCreateTime(),
		})
	}
	page := int(resp.GetPage())
	pageSize := int(resp.GetPageSize())
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return publicprofile.ListResult{
		List:       list,
		Pagination: pagination.CalcNestPagination(int(resp.GetTotal()), pageSize, page),
	}
}
