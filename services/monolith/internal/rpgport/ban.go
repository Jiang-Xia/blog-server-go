// Package rpgport 单体模式 RPG 能力端口（进程内调用，无需 gRPC）。
package rpgport

import (
	"context"

	"github.com/Jiang-Xia/blog-server-go/pkg/rpgsvc"
	rpgpunish "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/punishment"
)

type localBanChecker struct {
	p *rpgpunish.PunishmentService
}

// NewLocalBanChecker 用进程内 PunishmentService 实现 BanChecker。
func NewLocalBanChecker(p *rpgpunish.PunishmentService) rpgsvc.BanChecker {
	if p == nil {
		c, _ := rpgsvc.NewGRPCBanChecker("")
		return c
	}
	return localBanChecker{p: p}
}

func (l localBanChecker) AssertNotBanned(ctx context.Context, uid int) error {
	return l.p.AssertNotBanned(ctx, uid)
}
