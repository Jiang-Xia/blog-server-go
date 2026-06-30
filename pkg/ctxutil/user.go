// Package ctxutil 从 context 读取鉴权后写入的用户信息。
package ctxutil

import "context"

type ctxKey int

const (
	keyUserID ctxKey = iota + 1
	keyRoles
)

// RoleBrief JWT/权限中间件注入的角色摘要。
type RoleBrief struct {
	ID       int    `json:"id"`
	RoleName string `json:"roleName"`
}

// WithUserID 写入当前用户 ID。
func WithUserID(ctx context.Context, uid int) context.Context {
	return context.WithValue(ctx, keyUserID, uid)
}

// UserID 读取当前用户 ID；未登录返回 0。
func UserID(ctx context.Context) int {
	v, _ := ctx.Value(keyUserID).(int)
	return v
}

// WithRoles 写入角色列表。
func WithRoles(ctx context.Context, roles []RoleBrief) context.Context {
	return context.WithValue(ctx, keyRoles, roles)
}

// Roles 读取角色列表。
func Roles(ctx context.Context) []RoleBrief {
	v, _ := ctx.Value(keyRoles).([]RoleBrief)
	return v
}

// IsSuperAdmin 是否超级管理员（roleId=1）。
func IsSuperAdmin(ctx context.Context) bool {
	for _, r := range Roles(ctx) {
		if r.ID == 1 {
			return true
		}
	}
	return false
}
