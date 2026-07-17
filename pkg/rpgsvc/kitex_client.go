// Package rpgsvc RPG Kitex 客户端实现。
package rpgsvc

import (
	"context"
	"fmt"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/pkg/kitexreg"
	rpgv1 "github.com/Jiang-Xia/blog-server-go/proto/kitex/rpg/v1"
	"github.com/Jiang-Xia/blog-server-go/proto/kitex/rpg/v1/rpgservice"
	"github.com/cloudwego/kitex/client"
)

type kitexBanChecker struct {
	client rpgservice.Client
}

// NewKitexBanChecker 经 Nacos 发现 rpg-service；未配置时返回 noop（单体/测试跳过禁言校验）。
func NewKitexBanChecker(reg config.RegistryConfig) (BanChecker, error) {
	if !reg.Enabled() {
		return noopBanChecker{}, nil
	}
	r, err := kitexreg.NewResolver(reg)
	if err != nil {
		return nil, err
	}
	cli, err := rpgservice.NewClient(config.KitexServiceRPG, client.WithResolver(r))
	if err != nil {
		return nil, fmt.Errorf("new rpg kitex client: %w", err)
	}
	return &kitexBanChecker{client: cli}, nil
}

func (g *kitexBanChecker) AssertNotBanned(ctx context.Context, uid int) error {
	if uid <= 0 {
		return nil
	}
	resp, err := g.client.AssertNotBanned(ctx, &rpgv1.AssertNotBannedRequest{UserId: uint64(uid)})
	if err != nil {
		return errcode.WithMessage(errcode.InternalError, "禁言校验服务暂不可用")
	}
	if resp.GetBanned() {
		msg := resp.GetMessage()
		if msg == "" {
			msg = "您已被禁言"
		}
		return errcode.WithMessage(errcode.Forbidden, "%s", msg)
	}
	return nil
}

// noopBanChecker 未配置 Nacos 时跳过禁言校验（本地单体/测试）。
type noopBanChecker struct{}

func (noopBanChecker) AssertNotBanned(context.Context, int) error { return nil }
