// Package proxy gateway 反向代理：按路径前缀转发至 user/blog/rpg HTTP 服务。
package proxy

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/grpcmeta"
	"github.com/Jiang-Xia/blog-server-go/pkg/response"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/adaptor"
)

// Router 按附录 A 路径映射转发请求。
type Router struct {
	user        *httputil.ReverseProxy
	blog        *httputil.ReverseProxy
	rpg         *httputil.ReverseProxy
	profileBFF  app.HandlerFunc
}

// NewRouter 构造反向代理路由。
func NewRouter(cfg *config.Config) (*Router, error) {
	user, err := newProxy(cfg.Proxy.UserURL, "user")
	if err != nil {
		return nil, err
	}
	blog, err := newProxy(cfg.Proxy.BlogURL, "blog")
	if err != nil {
		return nil, err
	}
	rpg, err := newProxy(cfg.Proxy.RPGURL, "rpg")
	if err != nil {
		return nil, err
	}
	return &Router{user: user, blog: blog, rpg: rpg}, nil
}

func newProxy(raw, service string) (*httputil.ReverseProxy, error) {
	if raw == "" {
		return nil, nil
	}
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	p := httputil.NewSingleHostReverseProxy(u)
	p.ErrorHandler = func(w http.ResponseWriter, _ *http.Request, _ error) {
		writeUpstreamError(w, service)
	}
	return p, nil
}

// SetProfileBFF 注入 GET /user/public/:uid BFF 处理器（由 catch-all 在 pick 前分发）。
func (r *Router) SetProfileBFF(h app.HandlerFunc) {
	r.profileBFF = h
}

// ProxyHandler 返回将请求按 pick 规则转发的 Hertz 处理器（供 gateway 显式注册子路径）。
func (r *Router) ProxyHandler(apiPrefix string) app.HandlerFunc {
	return r.proxyHandler(apiPrefix)
}

// Register 挂载 catch-all 代理与 /realtime WS 转发。
func (r *Router) Register(h *server.Hertz, apiPrefix string) {
	h.Any("/realtime", adaptor.HertzHandler(r.wsHandler()))
	h.Any("/realtime/*path", adaptor.HertzHandler(r.wsHandler()))

	api := apiPrefix
	if api == "" {
		api = "/api/v1"
	}
	h.Any(api+"/*path", r.proxyHandler(api))
}

func (r *Router) wsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if r.blog != nil {
			r.blog.ServeHTTP(w, req)
		} else {
			writeUpstreamError(w, "blog")
		}
	})
}

func (r *Router) proxyHandler(apiPrefix string) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		path := string(c.Path())
		rel := strings.TrimPrefix(path, apiPrefix)
		rel = strings.TrimPrefix(rel, "/")
		if isPublicProfileBFF(rel) && r.profileBFF != nil {
			r.profileBFF(ctx, c)
			return
		}
		service, proxy := r.pick(path, apiPrefix)
		if proxy == nil {
			msg := "该接口未配置上游服务"
			if service != "" {
				msg = upstreamMessage(service)
			}
			c.JSON(http.StatusBadGateway, response.Body{
				Code:    upstreamUnavailableCode,
				BizCode: upstreamUnavailableCode,
				Message: msg,
			})
			return
		}
		if uid := c.GetHeader(grpcmeta.UserIDKey); len(uid) > 0 {
			c.Request.Header.Set(grpcmeta.UserIDKey, string(uid))
		}
		adaptor.HertzHandler(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			proxy.ServeHTTP(w, req)
		}))(ctx, c)
	}
}

func (r *Router) pick(path, apiPrefix string) (service string, proxy *httputil.ReverseProxy) {
	rel := strings.TrimPrefix(path, apiPrefix)
	rel = strings.TrimPrefix(rel, "/")

	if strings.HasPrefix(rel, "pub/") {
		return "", nil
	}
	if rel == "article/info" {
		return "", nil
	}
	if isPublicProfileBFF(rel) {
		return "", nil
	}
	if isRPGRoute(rel) {
		return "rpg", r.rpg
	}
	if isUserRoute(rel) {
		return "user", r.user
	}
	if isBlogRoute(rel) {
		return "blog", r.blog
	}
	if isRagRoute(rel) {
		return "blog", r.blog
	}
	return "blog", r.blog
}

func isRagRoute(rel string) bool {
	return rel == "rag" || strings.HasPrefix(rel, "rag/") ||
		rel == "admin/rag" || strings.HasPrefix(rel, "admin/rag/")
}

func isUserRoute(rel string) bool {
	prefixes := []string{
		"user", "captcha", "role", "dept", "privilege", "admin/menu",
		"sensitive-word", "operation-log",
	}
	for _, p := range prefixes {
		if rel == p || strings.HasPrefix(rel, p+"/") {
			return true
		}
	}
	if strings.HasPrefix(rel, "admin/") && !strings.HasPrefix(rel, "admin/rpg") && !strings.HasPrefix(rel, "admin/rag") {
		return true
	}
	return false
}

func isBlogRoute(rel string) bool {
	if rel == "scheduled-task" || strings.HasPrefix(rel, "scheduled-task/") {
		return true
	}
	return false
}

func isRPGRoute(rel string) bool {
	return strings.HasPrefix(rel, "rpg/") ||
		strings.HasPrefix(rel, "admin/rpg") ||
		strings.HasPrefix(rel, "pay") ||
		strings.HasPrefix(rel, "user/public/") ||
		strings.HasPrefix(rel, "rpg/public/")
}

// isPublicProfileBFF 精确匹配 user/public/:uid（不含 articles/collects/likes 子路径）。
func isPublicProfileBFF(rel string) bool {
	parts := strings.Split(rel, "/")
	if len(parts) != 3 || parts[0] != "user" || parts[1] != "public" {
		return false
	}
	_, err := strconv.Atoi(parts[2])
	return err == nil
}
