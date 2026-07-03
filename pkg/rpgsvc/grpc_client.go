// Package rpgsvc RPG gRPC 客户端实现。
package rpgsvc

import (
	"context"
	"fmt"

	rpgv1 "github.com/Jiang-Xia/blog-server-go/proto/gen/go/rpg/v1"
	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type grpcBanChecker struct {
	client rpgv1.RpgServiceClient
}

// NewGRPCBanChecker 连接 rpg-service gRPC 并返回 BanChecker。
func NewGRPCBanChecker(addr string) (BanChecker, error) {
	if addr == "" {
		return noopBanChecker{}, nil
	}
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("dial rpg grpc %s: %w", addr, err)
	}
	return &grpcBanChecker{client: rpgv1.NewRpgServiceClient(conn)}, nil
}

func (g *grpcBanChecker) AssertNotBanned(ctx context.Context, uid int) error {
	if uid <= 0 {
		return nil
	}
	resp, err := g.client.AssertNotBanned(ctx, &rpgv1.AssertNotBannedRequest{UserId: uint64(uid)})
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.Unavailable {
			return errcode.WithMessage(errcode.InternalError, "禁言校验服务暂不可用")
		}
		return err
	}
	if resp.GetBanned() {
		msg := resp.GetMessage()
		if msg == "" {
			msg = "您已被禁言"
		}
		return errcode.WithMessage(errcode.Forbidden, msg)
	}
	return nil
}

// noopBanChecker 未配置 rpg_addr 时跳过禁言校验（本地单体/测试）。
type noopBanChecker struct{}

func (noopBanChecker) AssertNotBanned(context.Context, int) error { return nil }
