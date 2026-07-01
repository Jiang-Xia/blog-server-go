// Package grpcclient gateway 连接内部微服务 gRPC。
package grpcclient

import (
	"fmt"
	"sync"

	blogv1 "github.com/Jiang-Xia/blog-server-go/proto/gen/go/blog/v1"
	rpgv1 "github.com/Jiang-Xia/blog-server-go/proto/gen/go/rpg/v1"
	userv1 "github.com/Jiang-Xia/blog-server-go/proto/gen/go/user/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Clients 聚合 user/blog/rpg gRPC 客户端（未配置地址时为 nil）。
type Clients struct {
	User userv1.UserServiceClient
	Blog blogv1.ArticleServiceClient
	RPG  rpgv1.RpgServiceClient
}

var (
	once    sync.Once
	loaded  *Clients
	loadErr error
)

// New 按配置地址建立 gRPC 连接（gateway 进程内单例）。
func New(userAddr, blogAddr, rpgAddr string) (*Clients, error) {
	once.Do(func() {
		c := &Clients{}
		if userAddr != "" {
			conn, err := grpc.NewClient(userAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				loadErr = fmt.Errorf("dial user grpc: %w", err)
				return
			}
			c.User = userv1.NewUserServiceClient(conn)
		}
		if blogAddr != "" {
			conn, err := grpc.NewClient(blogAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				loadErr = fmt.Errorf("dial blog grpc: %w", err)
				return
			}
			c.Blog = blogv1.NewArticleServiceClient(conn)
		}
		if rpgAddr != "" {
			conn, err := grpc.NewClient(rpgAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				loadErr = fmt.Errorf("dial rpg grpc: %w", err)
				return
			}
			c.RPG = rpgv1.NewRpgServiceClient(conn)
		}
		loaded = c
	})
	if loadErr != nil {
		return nil, loadErr
	}
	return loaded, nil
}
