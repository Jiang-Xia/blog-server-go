// Package jwtauth JWT 签发与校验，供 gateway 与各服务共享（对齐 Nest JwtModule）。
package jwtauth

import (
	"fmt"
	"time"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/golang-jwt/jwt/v5"
)

// Claims JWT payload，字段与 Nest certificate() 一致。
type Claims struct {
	ID       int           `json:"id"`
	Nickname string        `json:"nickname"`
	Username string        `json:"username"`
	Role     []RolePayload `json:"role"`
	jwt.RegisteredClaims
}

// RolePayload 写入 token 的角色摘要。
type RolePayload struct {
	ID       int    `json:"id"`
	RoleName string `json:"roleName"`
	RoleDesc string `json:"roleDesc,omitempty"`
}

// TokenTriple 兼容 Nest 三 token 响应。
type TokenTriple struct {
	Token        string `json:"token"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// Service 签发与解析 JWT。
type Service struct {
	secret     []byte
	legacyTTL  time.Duration
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewService 构造 JWT Service。
func NewService(cfg *config.Config) *Service {
	j := cfg.JWT
	legacy := j.LegacyTTL
	if legacy == 0 {
		legacy = 8 * time.Hour
	}
	access := j.AccessTTL
	if access == 0 {
		access = 30 * time.Minute
	}
	refresh := j.RefreshTTL
	if refresh == 0 {
		refresh = 7 * 24 * time.Hour
	}
	return &Service{
		secret:     []byte(j.Secret),
		legacyTTL:  legacy,
		accessTTL:  access,
		refreshTTL: refresh,
	}
}

// SignTriple 签发 legacy + access + refresh 三个 token。
func (j *Service) SignTriple(id int, nickname, username string, roles []RolePayload) (*TokenTriple, error) {
	legacy, err := j.sign(id, nickname, username, roles, j.legacyTTL)
	if err != nil {
		return nil, err
	}
	access, err := j.sign(id, nickname, username, roles, j.accessTTL)
	if err != nil {
		return nil, err
	}
	refresh, err := j.sign(id, nickname, username, roles, j.refreshTTL)
	if err != nil {
		return nil, err
	}
	return &TokenTriple{Token: legacy, AccessToken: access, RefreshToken: refresh}, nil
}

func (j *Service) sign(id int, nickname, username string, roles []RolePayload, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		ID:       id,
		Nickname: nickname,
		Username: username,
		Role:     roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(j.secret)
}

// Verify 解析并校验 token。
func (j *Service) Verify(token string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return j.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

// RemainingTTL 返回 token 剩余有效期（秒）。
func (j *Service) RemainingTTL(claims *Claims) int {
	if claims == nil || claims.ExpiresAt == nil {
		return 1
	}
	sec := int(time.Until(claims.ExpiresAt.Time).Seconds())
	if sec < 1 {
		return 1
	}
	return sec
}
