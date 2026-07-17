// Package blogsvc blog Kitex 客户端实现（审核状态同步等）。
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

type kitexModerationSyncer struct {
	client articleservice.Client
}

// NewKitexModerationSyncer 经 etcd 发现 blog-service；endpoints 为空时返回 noop（单体/测试）。
func NewKitexModerationSyncer(endpoints []string) (ContentModerationSyncer, error) {
	if len(endpoints) == 0 {
		return noopModerationSyncer{}, nil
	}
	r, err := kitexreg.NewResolver(endpoints)
	if err != nil {
		return nil, err
	}
	cli, err := articleservice.NewClient(config.KitexServiceBlog, client.WithResolver(r))
	if err != nil {
		return nil, fmt.Errorf("new blog kitex client: %w", err)
	}
	return &kitexModerationSyncer{client: cli}, nil
}

func (g *kitexModerationSyncer) UpdateContentModerationStatus(ctx context.Context, sourceType, sourceID, status string) error {
	_, err := g.client.UpdateContentModerationStatus(ctx, &blogv1.UpdateContentModerationStatusRequest{
		SourceType: sourceType,
		SourceId:   sourceID,
		Status:     status,
	})
	return err
}

type noopModerationSyncer struct{}

func (noopModerationSyncer) UpdateContentModerationStatus(context.Context, string, string, string) error {
	return nil
}
