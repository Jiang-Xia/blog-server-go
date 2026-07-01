// Package grpcserver 启动 blog-service gRPC 监听。
package grpcserver

import (
	"fmt"
	"net"

	"github.com/Jiang-Xia/blog-server-go/pkg/grpcmeta"
	blogv1 "github.com/Jiang-Xia/blog-server-go/proto/gen/go/blog/v1"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

// Run 启动 gRPC 监听；addr 为空时不启动。
func Run(addr string, srv *Server) (*grpc.Server, error) {
	if addr == "" {
		return nil, nil
	}
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("listen grpc %s: %w", addr, err)
	}
	gs := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.UnaryInterceptor(grpcmeta.AuthUnaryInterceptor()),
	)
	blogv1.RegisterArticleServiceServer(gs, srv)
	go func() {
		_ = gs.Serve(lis)
	}()
	return gs, nil
}
