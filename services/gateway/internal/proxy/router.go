// Package proxy gateway 反向代理：按路径前缀转发至 user/blog/rpg HTTP 服务。
package proxy

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/grpcmeta"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/adaptor"
)

// Router 按附录 A 路径映射转发请求。
type Router struct {
	user *httputil.ReverseProxy
	blog *httputil.ReverseProxy
	rpg  *httputil.ReverseProxy
}

// NewRouter 构造反向代理路由。
func NewRouter(cfg *config.Config) (*Router, error) {
	user, err := newProxy(cfg.Proxy.UserURL)
	if err != nil {
		return nil, err
	}
	blog, err := newProxy(cfg.Proxy.BlogURL)
	if err != nil {
		return nil, err
	}
	rpg, err := newProxy(cfg.Proxy.RPGURL)
	if err != nil {
		return nil, err
	}
	return &Router{user: user, blog: blog, rpg: rpg}, nil
}

func newProxy(raw string) (*httputil.ReverseProxy, error) {
	if raw == "" {
		return nil, nil
	}
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	return httputil.NewSingleHostReverseProxy(u), nil
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
			http.Error(w, "blog upstream not configured", http.StatusBadGateway)
		}
	})
}

func (r *Router) proxyHandler(apiPrefix string) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		path := string(c.Path())
		proxy := r.pick(path, apiPrefix)
		if proxy == nil {
			c.String(http.StatusBadGateway, "upstream not configured")
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

func (r *Router) pick(path, apiPrefix string) *httputil.ReverseProxy {
	rel := strings.TrimPrefix(path, apiPrefix)
	rel = strings.TrimPrefix(rel, "/")

	if strings.HasPrefix(rel, "pub/") {
		return nil
	}
	if isUserRoute(rel) {
		return r.user
	}
	if isRPGRoute(rel) {
		return r.rpg
	}
	return r.blog
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
	if strings.HasPrefix(rel, "admin/") && !strings.HasPrefix(rel, "admin/rpg") {
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
