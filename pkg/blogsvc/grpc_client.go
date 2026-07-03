// Package blogsvc blog gRPC 客户端实现。
package blogsvc

import (
	"context"
	"fmt"

	blogv1 "github.com/Jiang-Xia/blog-server-go/proto/gen/go/blog/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type grpcModerationSyncer struct {
	client blogv1.ArticleServiceClient
}

// NewGRPCModerationSyncer 连接 blog-service gRPC 并返回 ContentModerationSyncer。
func NewGRPCModerationSyncer(addr string) (ContentModerationSyncer, error) {
	if addr == "" {
		return noopModerationSyncer{}, nil
	}
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("dial blog grpc %s: %w", addr, err)
	}
	return &grpcModerationSyncer{client: blogv1.NewArticleServiceClient(conn)}, nil
}

func (g *grpcModerationSyncer) UpdateContentModerationStatus(ctx context.Context, sourceType, sourceID, status string) error {
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
