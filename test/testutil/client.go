// Package testutil 为 test/smoke、test/integration、test/e2e 提供 HTTP 客户端与断言辅助。
package testutil

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Jiang-Xia/blog-server-go/internal/devlogin"
	"github.com/Jiang-Xia/blog-server-go/pkg/jwtauth"
)

const defaultGatewayBase = "http://127.0.0.1:8000"

// APIResponse 与 Nest/Go 统一响应 {code, message, data} 对齐。
type APIResponse struct {
	Code    int             `json:"code"`
	BizCode int             `json:"bizCode"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// Client 本地 HTTP 测试客户端（仅 localhost）。
type Client struct {
	Origin  string // http://127.0.0.1:8000
	APIBase string // http://127.0.0.1:8000/api/v1
	HTTP    *http.Client
}

// NewClient 构造测试客户端；base 为空时读 TEST_BASE / DEV_LOGIN_BASE，默认 gateway :8000。
// 不依赖 configs/*.yaml（API 前缀默认 api/v1）；登录/JWT 签发时再加载配置。
func NewClient(base string) (*Client, error) {
	if base == "" {
		base = envOr("TEST_BASE", envOr("DEV_LOGIN_BASE", envOr("LOCAL_BASE", defaultGatewayBase)))
	}
	base = strings.TrimRight(base, "/")
	u, err := http.NewRequest(http.MethodGet, base+"/", nil)
	if err != nil {
		return nil, err
	}
	host := strings.ToLower(u.URL.Hostname())
	if host != "127.0.0.1" && host != "localhost" && host != "::1" {
		return nil, fmt.Errorf("拒绝非 localhost 目标: %s", base)
	}

	apiPrefix := "api/v1"
	if cfg, err := TryLoadConfig(); err == nil {
		if p := strings.Trim(cfg.App.APIPrefix, "/"); p != "" {
			apiPrefix = p
		}
	}
	return &Client{
		Origin:  base,
		APIBase: base + "/" + apiPrefix,
		HTTP:    &http.Client{Timeout: 20 * time.Second},
	}, nil
}

// Ping 探测 gateway health。
func (c *Client) Ping(ctx context.Context) bool {
	resp, _, err := c.Do(ctx, http.MethodGet, "/health", nil, "")
	return err == nil && resp != nil && resp.Code == 200
}

// RequireServer 服务不可达时 Skip（集成/E2E 前置）。
func RequireServer(t *testing.T, c *Client) {
	t.Helper()
	if !c.Ping(context.Background()) {
		t.Skip("gateway 未启动，请先 .\\scripts\\dev-all.ps1")
	}
}

// Do 发起请求；path 可为 /health 或完整 URL。
func (c *Client) Do(ctx context.Context, method, path string, body any, token string) (*APIResponse, int, error) {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, 0, err
		}
		reader = bytes.NewReader(b)
	}
	url := path
	if !strings.HasPrefix(path, "http") {
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		url = c.APIBase + path
	}
	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return nil, 0, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	res, err := c.HTTP.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	var api APIResponse
	if err := json.Unmarshal(raw, &api); err != nil {
		return nil, res.StatusCode, fmt.Errorf("parse json: %w body=%s", err, string(raw))
	}
	return &api, res.StatusCode, nil
}

// GET  shorthand。
func (c *Client) GET(ctx context.Context, path, token string) (*APIResponse, int, error) {
	return c.Do(ctx, http.MethodGet, path, nil, token)
}

// POST shorthand。
func (c *Client) POST(ctx context.Context, path string, body any, token string) (*APIResponse, int, error) {
	return c.Do(ctx, http.MethodPost, path, body, token)
}

// Login 超级管理员登录（devlogin）。
func (c *Client) Login(ctx context.Context) (string, error) {
	cfg, err := TryLoadConfig()
	if err != nil {
		return "", err
	}
	user := envOr("TEST_USERNAME", devlogin.DefaultUsername)
	pass := envOr("TEST_PASSWORD", devlogin.DefaultPassword)
	tokens, err := devlogin.Login(ctx, cfg, c.APIBase, user, pass)
	if err != nil {
		return "", err
	}
	return tokens.AccessToken, nil
}

// SignToken 为指定 uid 签发 JWT（集成测试多用户场景）。
func SignToken(uid int, username string) (string, error) {
	cfg, err := TryLoadConfig()
	if err != nil {
		return "", err
	}
	if cfg.JWT.Secret == "" {
		return "", fmt.Errorf("configs jwt.secret 为空，无法签发测试 token")
	}
	svc := jwtauth.NewService(cfg)
	triple, err := svc.SignTriple(uid, "test"+fmt.Sprint(uid), username, nil)
	if err != nil {
		return "", err
	}
	return triple.AccessToken, nil
}

// IsOK 业务成功（HTTP 200 且 code=200）。
func IsOK(resp *APIResponse) bool {
	return resp != nil && resp.Code == 200
}

// IsUnauthorized 未授权。
func IsUnauthorized(resp *APIResponse) bool {
	return resp != nil && resp.Code == 401
}

// UnmarshalData 将 data 解码到 dst。
func UnmarshalData(resp *APIResponse, dst any) error {
	if resp == nil {
		return fmt.Errorf("nil response")
	}
	return json.Unmarshal(resp.Data, dst)
}

// AssertOK 断言业务成功。
func AssertOK(t *testing.T, name string, resp *APIResponse, httpStatus int, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: request error: %v", name, err)
	}
	if httpStatus != http.StatusOK {
		t.Fatalf("%s: http status=%d", name, httpStatus)
	}
	if !IsOK(resp) {
		t.Fatalf("%s: code=%d message=%s", name, resp.Code, resp.Message)
	}
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
