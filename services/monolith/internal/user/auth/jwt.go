// Package auth 实现 JWT/OAuth/验证码；JWT 实现在 pkg/jwtauth。
package auth

import (
	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/jwtauth"
)

// Claims 兼容旧引用。
type Claims = jwtauth.Claims

// RolePayload 兼容旧引用。
type RolePayload = jwtauth.RolePayload

// TokenTriple 兼容旧引用。
type TokenTriple = jwtauth.TokenTriple

// JWTService 兼容旧引用。
type JWTService = jwtauth.Service

// NewJWTService 构造 JWTService。
func NewJWTService(cfg *config.Config) *JWTService {
	return jwtauth.NewService(cfg)
}
