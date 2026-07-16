// Package middleware 提供 Hertz 全局中间件：Recovery、RequestID、Logger、CORS。
package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/pkg/response"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ctxKeyRequestID struct{}

// RequestIDFromContext 从 context 读取请求 ID。
func RequestIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(ctxKeyRequestID{}).(string); ok {
		return v
	}
	return ""
}

// Recovery 捕获 panic 并返回统一 JSON。
func Recovery(log *zap.Logger) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		defer func() {
			if r := recover(); r != nil {
				log.Error("panic recovered",
					zap.Any("panic", r),
					zap.String("path", string(c.Path())),
					zap.String("request_id", RequestIDFromContext(ctx)),
				)
				response.Error(ctx, c, errcode.InternalError)
				c.Abort()
			}
		}()
		c.Next(ctx)
	}
}

// RequestID 为每个请求注入 X-Request-Id。
func RequestID() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		id := strings.TrimSpace(string(c.GetHeader("X-Request-Id")))
		if id == "" {
			id = uuid.NewString()
		}
		c.Response.Header.Set("X-Request-Id", id)
		ctx = context.WithValue(ctx, ctxKeyRequestID{}, id)
		c.Next(ctx)
	}
}

// Logger 记录请求方法、路径、耗时与状态码。
func Logger(log *zap.Logger) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		start := time.Now()
		c.Next(ctx)
		log.Info("http request",
			zap.String("method", string(c.Method())),
			zap.String("path", string(c.Path())),
			zap.Int("status", c.Response.StatusCode()),
			zap.Duration("latency", time.Since(start)),
			zap.String("request_id", RequestIDFromContext(ctx)),
		)
	}
}

// CORS 处理跨域，origins 为空时允许全部（开发环境）。
// 反射具体 Origin 时附带 Allow-Credentials，以支持管理端 withCredentials（验证码 Cookie）。
func CORS(cfg *config.Config) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		origin := strings.TrimSpace(string(c.GetHeader("Origin")))
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
			c.Response.Header.Set("Vary", "Origin")
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
