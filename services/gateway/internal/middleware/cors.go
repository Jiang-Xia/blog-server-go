// Package middleware gateway HTTP 中间件。
package middleware

import (
	"context"
	"net/http"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/cloudwego/hertz/pkg/app"
)

// CORS 处理跨域；origins 为空时允许全部（学习/本地联调）。
func CORS(cfg *config.Config) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		origin := string(c.GetHeader("Origin"))
		allowed := len(cfg.HTTP.CORSOrigins) == 0
		for _, o := range cfg.HTTP.CORSOrigins {
			if o == origin || o == "*" {
				allowed = true
				break
			}
		}
		if allowed && origin != "" {
			c.Response.Header.Set("Access-Control-Allow-Origin", origin)
			c.Response.Header.Set("Access-Control-Allow-Credentials", "true")
		} else if len(cfg.HTTP.CORSOrigins) == 0 {
			c.Response.Header.Set("Access-Control-Allow-Origin", "*")
		}
		c.Response.Header.Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Response.Header.Set("Access-Control-Allow-Headers", "Authorization,Content-Type,X-Request-Id")
		c.Response.Header.Set("Access-Control-Expose-Headers", "X-Request-Id")
		c.Response.Header.Set("Access-Control-Max-Age", "86400")
		if string(c.Method()) == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next(ctx)
	}
}
