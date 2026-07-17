// Package kitexclient gateway 经 Nacos 发现内部微服务 Kitex 客户端。
package kitexclient

import (
	"fmt"
	"sync"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/kitexreg"
	"github.com/Jiang-Xia/blog-server-go/proto/kitex/blog/v1/articleservice"
	"github.com/Jiang-Xia/blog-server-go/proto/kitex/rpg/v1/rpgservice"
	"github.com/Jiang-Xia/blog-server-go/proto/kitex/user/v1/userservice"
	"github.com/cloudwego/kitex/client"
)

// Clients 聚合 user/blog/rpg Kitex 客户端。
type Clients struct {
	User userservice.Client
	Blog articleservice.Client
	RPG  rpgservice.Client
}

var (
	once    sync.Once
	loaded  *Clients
	loadErr error
)

// New 按 Nacos 配置建立 Kitex 客户端（gateway 进程内单例）。
func New(reg config.RegistryConfig) (*Clients, error) {
	once.Do(func() {
		if !reg.Enabled() {
			loadErr = fmt.Errorf("registry.nacos_addr required for gateway Kitex clients")
			return
		}
		r, err := kitexreg.NewResolver(reg)
		if err != nil {
			loadErr = err
			return
		}
		userCli, err := userservice.NewClient(config.KitexServiceUser, client.WithResolver(r))
		if err != nil {
			loadErr = fmt.Errorf("new user kitex client: %w", err)
			return
		}
		blogCli, err := articleservice.NewClient(config.KitexServiceBlog, client.WithResolver(r))
		if err != nil {
			loadErr = fmt.Errorf("new blog kitex client: %w", err)
			return
		}
		rpgCli, err := rpgservice.NewClient(config.KitexServiceRPG, client.WithResolver(r))
		if err != nil {
			loadErr = fmt.Errorf("new rpg kitex client: %w", err)
			return
		}
		loaded = &Clients{User: userCli, Blog: blogCli, RPG: rpgCli}
	})
	if loadErr != nil {
		return nil, loadErr
	}
	return loaded, nil
}
