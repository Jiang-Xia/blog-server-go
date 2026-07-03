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
	tongjiTokenRedisKey = "baidu:tongji:access_token"
	tongjiTokenTTLSec   = 25 * 24 * 3600
	tongjiAPIPrefix     = "/rest/2.0/tongji/"
)

// TongjiRefresher 百度统计 token 刷新（scheduled-task 与 resources 共用）。
type TongjiRefresher interface {
	ForceRefreshTongjiAccessToken(ctx context.Context) (map[string]interface{}, error)
}

// BaiduTongJi 代理百度统计 OpenAPI（token 缺失或过期时自动 refresh 后重试）。
func (s *ResourcesService) BaiduTongJi(ctx context.Context, query map[string]string) (interface{}, error) {
	normalized, err := normalizeTongjiQuery(query)
	if err != nil {
		return nil, err
	}
	data, err := s.getTongjiData(ctx, normalized)
	if err != nil {
		if isTongjiTokenMissingErr(err) {
			if _, refreshErr := s.RefreshAccessToken(ctx); refreshErr != nil {
				return nil, refreshErr
			}
			return s.getTongjiData(ctx, normalized)
		}
		return nil, err
	}
	if isBaiduTokenExpired(data) {
		if _, refreshErr := s.RefreshAccessToken(ctx); refreshErr != nil {
			return nil, refreshErr
		}
		return s.getTongjiData(ctx, normalized)
	}
	return data, nil
}

// ForceRefreshTongjiAccessToken 清除缓存并重新换取 access_token（超管手动刷新）。
func (s *ResourcesService) ForceRefreshTongjiAccessToken(ctx context.Context) (map[string]interface{}, error) {
	s.tongjiMu.Lock()
	s.tongjiMemCache = ""
	s.tongjiMu.Unlock()
	if s.redis != nil {
		_ = s.redis.Del(ctx, tongjiTokenRedisKey)
	}
	if _, err := s.RefreshAccessToken(ctx); err != nil {
		return nil, err
	}
	return map[string]interface{}{"refreshed": true}, nil
}

// RefreshAccessToken 用 refresh_token 换取 access_token 并写入 Redis。
func (s *ResourcesService) RefreshAccessToken(ctx context.Context) (map[string]interface{}, error) {
	cfg := s.cfg.App
	if cfg.TongjiRefreshToken == "" || cfg.TongjiClientID == "" || cfg.TongjiClientSecret == "" {
		return nil, errcode.WithMessage(errcode.InternalError,
			"百度统计 OAuth 配置不完整，请检查 app_tongjiRefreshToken / app_tongjiClientId / app_tongjiClientSecret")
	}
	tokenURL := "http://openapi.baidu.com/oauth/2.0/token?" + url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {cfg.TongjiRefreshToken},
		"client_id":     {cfg.TongjiClientID},
		"client_secret": {cfg.TongjiClientSecret},
	}.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, tokenURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, errcode.WithMessage(errcode.InternalError, "刷新百度统计 access_token 失败: %s", err.Error())
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, errcode.WithMessage(errcode.InternalError, "刷新百度统计 access_token 失败: 响应解析错误")
	}
	accessToken, _ := parsed["access_token"].(string)
	if accessToken == "" {
		msg := tongjiErrMessage(parsed)
		return nil, errcode.WithMessage(errcode.InternalError, "刷新百度统计 access_token 失败: %s", msg)
	}
	s.tongjiMu.Lock()
	s.tongjiMemCache = accessToken
	s.tongjiMu.Unlock()
	if s.redis != nil {
		if err := s.redis.Set(ctx, tongjiTokenRedisKey, accessToken, tongjiTokenTTLSec); err != nil {
			return nil, err
		}
	}
	return parsed, nil
}

func (s *ResourcesService) getTongjiData(ctx context.Context, query map[string]string) (map[string]interface{}, error) {
	apiPath := query["url"]
	accessToken, err := s.readTongjiAccessToken(ctx)
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	for k, v := range query {
		if k == "url" {
			continue
		}
		params.Set(k, v)
	}
	params.Set("access_token", accessToken)

	reqURL := "https://openapi.baidu.com" + apiPath + "?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, errcode.WithMessage(errcode.InternalError, "百度统计接口调用失败: %s", err.Error())
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, errcode.WithMessage(errcode.InternalError, "百度统计接口调用失败: 响应解析错误")
	}
	if msg := strings.TrimSpace(fmt.Sprint(data["error_description"])); msg != "" && data["access_token"] == nil {
		return nil, errcode.WithMessage(errcode.InternalError, "百度统计接口调用失败: %s", msg)
	}
	return data, nil
}

func (s *ResourcesService) readTongjiAccessToken(ctx context.Context) (string, error) {
	s.tongjiMu.RLock()
	if s.tongjiMemCache != "" {
		token := s.tongjiMemCache
		s.tongjiMu.RUnlock()
		return token, nil
	}
	s.tongjiMu.RUnlock()

	if s.redis != nil {
		token, err := s.redis.Get(ctx, tongjiTokenRedisKey)
		if err != nil {
			return "", err
		}
		if token != "" {
			s.tongjiMu.Lock()
			s.tongjiMemCache = token
			s.tongjiMu.Unlock()
			return token, nil
		}
	}
	return "", errcode.WithMessage(errcode.InternalError, "百度统计 access_token 不存在，请先触发 token 刷新")
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

func isBaiduTokenExpired(data map[string]interface{}) bool {
	if data == nil {
		return false
	}
	switch code := data["error_code"].(type) {
	case float64:
		return int(code) == 110 || int(code) == 111
	case int:
		return code == 110 || code == 111
	case json.Number:
		n, _ := code.Int64()
		return n == 110 || n == 111
	default:
		return false
	}
}

func isTongjiTokenMissingErr(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "access_token 不存在")
}

func tongjiErrMessage(parsed map[string]interface{}) string {
	if msg, ok := parsed["error_description"].(string); ok && msg != "" {
		return msg
	}
	if msg, ok := parsed["error"].(string); ok && msg != "" {
		return msg
	}
	return "百度返回的 access_token 为空"
}

var _ TongjiRefresher = (*ResourcesService)(nil)
