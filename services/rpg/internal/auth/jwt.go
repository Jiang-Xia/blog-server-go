// Package auth JWT 签发与校验封装（底层 pkg/jwtauth）。
package auth

import (
	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/jwtauth"
)

// Claims 兼容 handler/middleware 引用。
type Claims = jwtauth.Claims

// RolePayload 兼容 handler/middleware 引用。
type RolePayload = jwtauth.RolePayload

// JWTService 兼容 handler/middleware 引用。
type JWTService = jwtauth.Service

// NewJWTService 构造 JWTService。
func NewJWTService(cfg *config.Config) *JWTService {
	return jwtauth.NewService(cfg)
}
