// Package kitexreg 封装 etcd Registry / Resolver，供学习路径微服务注册与发现。
package kitexreg

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/cloudwego/kitex/pkg/discovery"
	"github.com/cloudwego/kitex/pkg/registry"
	etcd "github.com/kitex-contrib/registry-etcd"
)

// NewRegistry 创建 etcd 服务注册器。
func NewRegistry(endpoints []string) (registry.Registry, error) {
	if len(endpoints) == 0 {
		return nil, fmt.Errorf("registry.etcd_endpoints required")
	}
	r, err := etcd.NewEtcdRegistry(endpoints)
	if err != nil {
		return nil, fmt.Errorf("new etcd registry: %w", err)
	}
	return r, nil
}

// NewResolver 创建 etcd 服务发现器。
func NewResolver(endpoints []string) (discovery.Resolver, error) {
	if len(endpoints) == 0 {
		return nil, fmt.Errorf("registry.etcd_endpoints required")
	}
	r, err := etcd.NewEtcdResolver(endpoints)
	if err != nil {
		return nil, fmt.Errorf("new etcd resolver: %w", err)
	}
	return r, nil
}

// ResolveServiceTCPAddr 解析 Kitex 监听/注册地址。
// 本机以 ":port" 配置时强制 127.0.0.1，避免 Windows 选中 APIPA/虚拟网卡导致客户端无法拨号；
// Docker 容器内（存在 /.dockerenv）保持 ":port"，由运行时选择容器网卡 IP。
func ResolveServiceTCPAddr(addr string) (*net.TCPAddr, error) {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return nil, fmt.Errorf("kitex.addr required")
	}
	if strings.HasPrefix(addr, ":") {
		if _, err := os.Stat("/.dockerenv"); err != nil {
			addr = "127.0.0.1" + addr
		}
	}
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("resolve kitex addr %s: %w", addr, err)
	}
	return tcpAddr, nil
}
