// Package kitexserver 启动 blog-service Kitex 监听并注册到 etcd。
package kitexserver

import (
	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/kitexmeta"
	"github.com/Jiang-Xia/blog-server-go/pkg/kitexreg"
	"github.com/Jiang-Xia/blog-server-go/proto/kitex/blog/v1/articleservice"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
)

// Run 启动 Kitex 并注册到 etcd；addr 或 etcd 为空时不启动。
func Run(cfg *config.Config, srv *Server) (server.Server, error) {
	addr := cfg.Kitex.Addr
	endpoints := cfg.Registry.EtcdEndpointsOrEmpty()
	if addr == "" || len(endpoints) == 0 {
		return nil, nil
	}
	r, err := kitexreg.NewRegistry(endpoints)
	if err != nil {
		return nil, err
	}
	tcpAddr, err := kitexreg.ResolveServiceTCPAddr(addr)
	if err != nil {
		return nil, err
	}
	svr := articleservice.NewServer(srv,
		server.WithServiceAddr(tcpAddr),
		server.WithRegistry(r),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: config.KitexServiceBlog}),
		server.WithMiddleware(kitexmeta.AuthMiddleware),
	)
	go func() {
		if err := svr.Run(); err != nil {
			_ = err
		}
	}()
	return svr, nil
}
