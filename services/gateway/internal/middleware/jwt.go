// Package middleware gateway JWT 验签并透传 x-user-id。
package middleware

import (
	"context"
	"strconv"
	"strings"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/grpcmeta"
	"github.com/Jiang-Xia/blog-server-go/pkg/jwtauth"
	"github.com/cloudwego/hertz/pkg/app"
)

// JWTPassthrough 解析 Bearer JWT 并写入 x-user-id 请求头供下游使用。
func JWTPassthrough(cfg *config.Config) app.HandlerFunc {
	jwt := jwtauth.NewService(cfg)
	return func(ctx context.Context, c *app.RequestContext) {
		token := bearerToken(c)
		if token != "" {
			if claims, err := jwt.Verify(token); err == nil && claims.ID > 0 {
				c.Request.Header.Set(grpcmeta.UserIDKey, strconv.Itoa(claims.ID))
			}
		}
		c.Next(ctx)
	}
}

func bearerToken(c *app.RequestContext) string {
	authz := strings.TrimSpace(string(c.GetHeader("Authorization")))
	if authz == "" {
		return ""
	}
	const prefix = "Bearer "
	if strings.HasPrefix(authz, prefix) {
		return strings.TrimSpace(authz[len(prefix):])
	}
	return authz
}
