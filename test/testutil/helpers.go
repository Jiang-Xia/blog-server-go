package testutil

import (
	"context"
	"fmt"
	"net/http"
	"testing"
)

// MustLogin 登录并返回 token，失败则 Skip。
func MustLogin(t *testing.T, c *Client) string {
	t.Helper()
	token, err := c.Login(context.Background())
	return SkipUnlessLogin(t, c, token, err)
}

// MustSignToken 签发指定 uid 的 JWT，失败则 Skip。
func MustSignToken(t *testing.T, uid int, username string) string {
	t.Helper()
	token, err := SignToken(uid, username)
	if err != nil {
		t.Skipf("sign token uid=%d: %v", uid, err)
	}
	return token
}

// FirstArticleID 从 C 端文章列表取第一篇 id；无数据则 Skip。
func FirstArticleID(t *testing.T, c *Client, ctx context.Context, token string) int {
	t.Helper()
	resp, status, err := c.POST(ctx, "/article/list", map[string]any{
		"page": 1, "pageSize": 5, "client": true,
	}, token)
	AssertOK(t, "article/list for id", resp, status, err)
	var data struct {
		List []struct {
			ID int `json:"id"`
		} `json:"list"`
	}
	if err := UnmarshalData(resp, &data); err != nil || len(data.List) == 0 {
		t.Skip("无文章数据，请先 sync-data")
	}
	return data.List[0].ID
}

// AssertGETOK 批量 GET 并断言 code=200。
func AssertGETOK(t *testing.T, c *Client, ctx context.Context, token string, paths ...string) {
	t.Helper()
	for _, path := range paths {
		resp, status, err := c.GET(ctx, path, token)
		AssertOK(t, path, resp, status, err)
	}
}

// AssertPOSTOK 批量 POST（空 body）并断言 code=200。
func AssertPOSTOK(t *testing.T, c *Client, ctx context.Context, token string, paths ...string) {
	t.Helper()
	for _, path := range paths {
		resp, status, err := c.POST(ctx, path, nil, token)
		AssertOK(t, path, resp, status, err)
	}
}

// AssertUnauthorized POST 无 token 应返回 401。
func AssertUnauthorizedPOST(t *testing.T, c *Client, ctx context.Context, path string, body any) {
	t.Helper()
	resp, status, err := c.POST(ctx, path, body, "")
	if err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	if status != http.StatusOK || !IsUnauthorized(resp) {
		t.Fatalf("%s unauth want code=401, got http=%d code=%d", path, status, resp.Code)
	}
}

// AssertBizError POST/GET 期望业务错误（非 200 code，但 HTTP 200）。
func AssertBizError(t *testing.T, c *Client, ctx context.Context, method, path string, body any, token string, wantCode int) {
	t.Helper()
	var resp *APIResponse
	var status int
	var err error
	switch method {
	case http.MethodPost:
		resp, status, err = c.POST(ctx, path, body, token)
	case http.MethodGet:
		resp, status, err = c.GET(ctx, path, token)
	case http.MethodPatch:
		resp, status, err = c.PATCH(ctx, path, body, token)
	default:
		resp, status, err = c.Do(ctx, method, path, body, token)
	}
	if err != nil || status != http.StatusOK {
		t.Fatalf("%s %s: http=%d err=%v", method, path, status, err)
	}
	if resp.Code != wantCode {
		t.Fatalf("%s %s want code=%d, got %d msg=%s", method, path, wantCode, resp.Code, resp.Message)
	}
}

// PATCH shorthand。
func (c *Client) PATCH(ctx context.Context, path string, body any, token string) (*APIResponse, int, error) {
	return c.Do(ctx, http.MethodPatch, path, body, token)
}

// DELETE shorthand。
func (c *Client) DELETE(ctx context.Context, path, token string) (*APIResponse, int, error) {
	return c.Do(ctx, http.MethodDelete, path, nil, token)
}

// Pathf 格式化路径（供测试用）。
func Pathf(format string, args ...any) string {
	return fmt.Sprintf(format, args...)
}
