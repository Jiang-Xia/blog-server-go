package testutil

import (
	"os"
	"testing"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
)

// TryLoadConfig 依次尝试 CONFIG_PATH、gateway/user/monolith yaml。
func TryLoadConfig() (*config.Config, error) {
	if p := os.Getenv("CONFIG_PATH"); p != "" {
		return config.MustLoad(p)
	}
	for _, path := range []string{
		"configs/gateway.yaml",
		"configs/user.yaml",
		"configs/monolith.yaml",
	} {
		cfg, err := config.MustLoad(path)
		if err == nil {
			return cfg, nil
		}
	}
	return nil, os.ErrNotExist
}

// RequireConfig 无本地 yaml 时 Skip。
func RequireConfig(t *testing.T) *config.Config {
	t.Helper()
	cfg, err := TryLoadConfig()
	if err != nil {
		t.Skip("未找到 configs/*.yaml，请先 .\\scripts\\setup-config.ps1")
	}
	return cfg
}

// SkipUnlessLogin 登录失败时 Skip（服务未启或 Redis/账号不可用）。
func SkipUnlessLogin(t *testing.T, c *Client, token string, err error) string {
	t.Helper()
	if err != nil {
		t.Skipf("登录跳过: %v", err)
	}
	return token
}
