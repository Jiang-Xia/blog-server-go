// Package middleware JWT 鉴权：可选解析与强制校验 Bearer token。
package middleware

import (
	"context"
	"strings"

	"github.com/Jiang-Xia/blog-server-go/pkg/ctxutil"
	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/pkg/response"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/auth"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/repo"
	"github.com/cloudwego/hertz/pkg/app"
)

// OptionalJWT 若请求携带 Bearer token 则校验并写入 uid/roles；缺失或无效时不中断（供全局 Permission 前置解析）。
func OptionalJWT(jwtSvc *auth.JWTService) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		ctx = injectJWTClaims(ctx, c, jwtSvc)
		c.Next(ctx)
	}
}

// RequiredJWT 强制 Bearer 鉴权，对齐 Nest JwtAuthGuard（含锁定账号拦截）。
func RequiredJWT(jwtSvc *auth.JWTService, userRepo *repo.UserRepo) app.HandlerFunc {
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
		u, err := userRepo.FindByID(ctx, claims.ID)
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
		ctx = ctxutil.WithUserID(ctx, claims.ID)
		ctx = ctxutil.WithRoles(ctx, roleBriefsFromClaims(claims))
		c.Next(ctx)
	}
}

func injectJWTClaims(ctx context.Context, c *app.RequestContext, jwtSvc *auth.JWTService) context.Context {
	token := bearerToken(c)
	if token == "" {
		return ctx
	}
	claims, err := jwtSvc.Verify(token)
	if err != nil {
		return ctx
	}
	ctx = ctxutil.WithUserID(ctx, claims.ID)
	return ctxutil.WithRoles(ctx, roleBriefsFromClaims(claims))
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
