// Package kitexreg 封装 Nacos Registry / Resolver，供学习路径微服务注册与发现。
package kitexreg

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/cloudwego/kitex/pkg/discovery"
	"github.com/cloudwego/kitex/pkg/registry"
	nacosregistry "github.com/kitex-contrib/registry-nacos/registry"
	nacosresolver "github.com/kitex-contrib/registry-nacos/resolver"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

// NewRegistry 创建 Nacos 服务注册器。
func NewRegistry(reg config.RegistryConfig) (registry.Registry, error) {
	cli, err := newNamingClient(reg)
	if err != nil {
		return nil, err
	}
	return nacosregistry.NewNacosRegistry(cli, registryOpts(reg)...), nil
}

// NewResolver 创建 Nacos 服务发现器。
func NewResolver(reg config.RegistryConfig) (discovery.Resolver, error) {
	cli, err := newNamingClient(reg)
	if err != nil {
		return nil, err
	}
	return nacosresolver.NewNacosResolver(cli, resolverOpts(reg)...), nil
}

func registryOpts(reg config.RegistryConfig) []nacosregistry.Option {
	g := strings.TrimSpace(reg.Group)
	if g == "" {
		return nil
	}
	return []nacosregistry.Option{nacosregistry.WithGroup(g)}
}

func resolverOpts(reg config.RegistryConfig) []nacosresolver.Option {
	g := strings.TrimSpace(reg.Group)
	if g == "" {
		return nil
	}
	return []nacosresolver.Option{nacosresolver.WithGroup(g)}
}

func newNamingClient(reg config.RegistryConfig) (naming_client.INamingClient, error) {
	addr := strings.TrimSpace(reg.NacosAddr)
	if addr == "" {
		return nil, fmt.Errorf("registry.nacos_addr required")
	}
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("parse registry.nacos_addr %q: %w", addr, err)
	}
	port, err := strconv.ParseUint(portStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse nacos port in %q: %w", addr, err)
	}
	ns := strings.TrimSpace(reg.NamespaceID)
	// Nacos 保留命名空间 public 的 ID 为空字符串；配置写 public 时做兼容。
	if ns == "public" {
		ns = ""
	}
	tmp := filepath.Join(os.TempDir(), "nacos-kitex")
	cc := constant.ClientConfig{
		NamespaceId:         ns,
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              filepath.Join(tmp, "log"),
		CacheDir:            filepath.Join(tmp, "cache"),
		LogLevel:            "warn",
		Username:            strings.TrimSpace(reg.Username),
		Password:            strings.TrimSpace(reg.Password),
	}
	cli, err := clients.NewNamingClient(vo.NacosClientParam{
		ClientConfig:  &cc,
		ServerConfigs: []constant.ServerConfig{*constant.NewServerConfig(host, port)},
	})
	if err != nil {
		return nil, fmt.Errorf("new nacos naming client: %w", err)
	}
	return cli, nil
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
