package blogsvc

import (
	"context"
	"fmt"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/kitexreg"
	"github.com/Jiang-Xia/blog-server-go/pkg/pagination"
	"github.com/Jiang-Xia/blog-server-go/pkg/publicprofile"
	blogv1 "github.com/Jiang-Xia/blog-server-go/proto/kitex/blog/v1"
	"github.com/Jiang-Xia/blog-server-go/proto/kitex/blog/v1/articleservice"
	"github.com/cloudwego/kitex/client"
)

// PublicProfileLister 公开主页收藏/点赞列表（blog Kitex）。
type PublicProfileLister interface {
	ListPublicCollectArticles(ctx context.Context, uid, page, pageSize int) (publicprofile.ListResult, error)
	ListPublicLikeArticles(ctx context.Context, uid, page, pageSize int) (publicprofile.ListResult, error)
}

type kitexPublicProfileLister struct {
	client articleservice.Client
}

// NewKitexPublicProfileLister 经 etcd 发现 blog-service；endpoints 为空时返回错误。
func NewKitexPublicProfileLister(endpoints []string) (PublicProfileLister, error) {
	if len(endpoints) == 0 {
		return nil, fmt.Errorf("registry.etcd_endpoints required for rpg public profile lists")
	}
	r, err := kitexreg.NewResolver(endpoints)
	if err != nil {
		return nil, err
	}
	cli, err := articleservice.NewClient(config.KitexServiceBlog, client.WithResolver(r))
	if err != nil {
		return nil, fmt.Errorf("new blog kitex client: %w", err)
	}
	return &kitexPublicProfileLister{client: cli}, nil
}

func (g *kitexPublicProfileLister) ListPublicCollectArticles(ctx context.Context, uid, page, pageSize int) (publicprofile.ListResult, error) {
	resp, err := g.client.ListPublicCollectArticles(ctx, &blogv1.ListPublicProfileArticlesRequest{
		Uid: uint64(uid), Page: int32(page), PageSize: int32(pageSize),
	})
	if err != nil {
		return publicprofile.ListResult{}, err
	}
	return protoToListResult(resp), nil
}

func (g *kitexPublicProfileLister) ListPublicLikeArticles(ctx context.Context, uid, page, pageSize int) (publicprofile.ListResult, error) {
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
			"id":            int(item.GetId()),
			"title":         item.GetTitle(),
			"description":   item.GetDescription(),
			"cover":         item.GetCover(),
			"views":         int(item.GetViews()),
			"likes":         int(item.GetLikes()),
			"articleLevel":  int(item.GetArticleLevel()),
			"isMasterpiece": int(item.GetIsMasterpiece()),
			"tipTotal":      int(item.GetTipTotal()),
			"createTime":    item.GetCreateTime(),
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
