// Package middleware JWT 鉴权：强制校验 Bearer token（账号状态经 user gRPC）。
package middleware

import (
	"context"
	"strings"

	"github.com/Jiang-Xia/blog-server-go/pkg/ctxutil"
	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/pkg/response"
	"github.com/Jiang-Xia/blog-server-go/pkg/usersvc"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/auth"
	"github.com/cloudwego/hertz/pkg/app"
)

// RequiredJWT 强制 Bearer 鉴权，对齐 Nest JwtAuthGuard（含锁定账号拦截）。
func RequiredJWT(jwtSvc *auth.JWTService, users usersvc.UserService) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		token := bearerToken(c)
		if token == "" {
			response.Error(ctx, c, errcode.WithMessage(errcode.Unauthorized, "身份验证失败"))
			c.Abort()
			return
		}
		claims, err := jwtSvc.Verify(token)
		if err != nil {
			response.Error(ctx, c, errcode.WithMessage(errcode.Unauthorized, "身份验证失败"))
			c.Abort()
			return
		}
		if users != nil {
			u, err := users.GetUser(ctx, uint64(claims.ID))
			if err != nil {
				response.Error(ctx, c, errcode.WithMessage(errcode.Unauthorized, "身份验证失败"))
				c.Abort()
				return
			}
			if u.Status == "locked" {
				response.Error(ctx, c, errcode.WithMessage(errcode.Unauthorized, "账号已被锁定！"))
				c.Abort()
				return
			}
		}
		ctx = ctxutil.WithUserID(ctx, claims.ID)
		ctx = ctxutil.WithRoles(ctx, roleBriefsFromClaims(claims))
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

func roleBriefsFromClaims(claims *auth.Claims) []ctxutil.RoleBrief {
	if claims == nil || len(claims.Role) == 0 {
		return nil
	}
	out := make([]ctxutil.RoleBrief, 0, len(claims.Role))
	for _, r := range claims.Role {
		out = append(out, ctxutil.RoleBrief{ID: r.ID, RoleName: r.RoleName})
	}
	return out
}
