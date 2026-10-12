package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
)

const (
	tongjiAPIPrefix = "/rest/2.0/tongji/"
)

// TongjiRefresher 百度统计 token 刷新入口（scheduled-task 与 resources 共用）。
// 代理模式下 token 由 blog-server（Nest）统一管理，本服务不再持有百度 OAuth 凭据。
type TongjiRefresher interface {
	ForceRefreshTongjiAccessToken(ctx context.Context) (map[string]interface{}, error)
}

// BaiduTongJi 代理百度统计 OpenAPI：透传查询参数到 blog-server（Nest）的同名代理接口。
// token（access_token / refresh_token）由 Nest 单点管理，避免多服务共用同一条
// refresh_token 轮换链互相作废（百度 refresh_token 一次性，轮换即废）。
func (s *ResourcesService) BaiduTongJi(ctx context.Context, query map[string]string) (interface{}, error) {
	normalized, err := normalizeTongjiQuery(query)
	if err != nil {
		return nil, err
	}
	return s.proxyTongjiToNest(ctx, normalized)
}

// ForceRefreshTongjiAccessToken 代理模式下的刷新入口。
// token 归 Nest 管理，此处不做实际刷新，仅返回提示（保留接口以满足 TongjiRefresher 依赖）。
// 需要真正刷新时应在 blog-server 管理端触发：POST /scheduled-task/cache/refresh-tongji-token。
func (s *ResourcesService) ForceRefreshTongjiAccessToken(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{
		"refreshed":  false,
		"managed_by": "blog-server",
		"message":    "百度统计 token 已由 blog-server(Nest) 统一管理，请在其管理端刷新",
	}, nil
}

// proxyTongjiToNest 将查询参数透传给 Nest 的 /resources/baidutongji 并解包 data。
func (s *ResourcesService) proxyTongjiToNest(ctx context.Context, query map[string]string) (interface{}, error) {
	base := s.cfg.App.TongjiProxyBaseURLOrDefault()
	params := url.Values{}
	for k, v := range query {
		params.Set(k, v)
	}
	reqURL := base + "/resources/baidutongji?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, errcode.WithMessage(errcode.InternalError, "百度统计代理请求构造失败: %s", err.Error())
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, errcode.WithMessage(errcode.InternalError,
			"百度统计代理请求失败(检查 blog-server 是否在线): %s", err.Error())
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// Nest 统一响应格式：{code, bizCode, data, message}
	var envelope struct {
		Code    int         `json:"code"`
		BizCode int         `json:"bizCode"`
		Data    interface{} `json:"data"`
		Message string      `json:"message"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, errcode.WithMessage(errcode.InternalError,
			"百度统计代理响应解析错误(HTTP %d)", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK || envelope.Code != 200 {
		msg := strings.TrimSpace(envelope.Message)
		if msg == "" {
			msg = fmt.Sprintf("blog-server 代理返回 HTTP %d", resp.StatusCode)
		}
		return nil, errcode.WithMessage(errcode.InternalError, "百度统计接口调用失败: %s", msg)
	}
	return envelope.Data, nil
}

func normalizeTongjiQuery(query map[string]string) (map[string]string, error) {
	if query == nil {
		query = map[string]string{}
	}
	apiURL := strings.TrimSpace(query["url"])
	if apiURL == "" {
		return nil, errcode.WithMessage(errcode.InvalidParam, "缺少百度统计请求参数 url")
	}
	if !strings.HasPrefix(apiURL, tongjiAPIPrefix) {
		return nil, errcode.WithMessage(errcode.InvalidParam, "百度统计请求 url 不合法")
	}
	out := make(map[string]string, len(query))
	for k, v := range query {
		out[k] = v
	}
	out["url"] = apiURL
	return out, nil
}

var _ TongjiRefresher = (*ResourcesService)(nil)
