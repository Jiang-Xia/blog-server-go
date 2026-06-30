// Package vo 用户响应视图，脱敏与 JWT 角色摘要。
package vo

import (
	"time"

	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/repo"
)

// JWTRolePayload 写入 JWT 的角色摘要，与 Nest certificate() payload.role 对齐。
type JWTRolePayload struct {
	ID       int    `json:"id"`
	RoleName string `json:"roleName"`
	RoleDesc string `json:"roleDesc,omitempty"`
}

// SanitizeUser 将 Ent 用户转为 map，剔除 password/salt 等敏感字段。
func SanitizeUser(u *ent.User) map[string]interface{} {
	if u == nil {
		return nil
	}
	m := map[string]interface{}{
		"id":         u.ID,
		"nickname":   u.Nickname,
		"status":     u.Status,
		"intro":      u.Intro,
		"avatar":     u.Avatar,
		"homepage":   u.Homepage,
		"createTime": formatTime(u.CreateTime),
		"updateTime": formatTime(u.UpdateTime),
		"isDelete":   u.IsDelete,
		"version":    u.Version,
	}
	if u.Username != nil {
		m["username"] = *u.Username
	}
	if u.Email != nil {
		m["email"] = *u.Email
	}
	if u.GithubId != nil {
		m["githubId"] = *u.GithubId
	}
	if u.WechatOpenId != nil {
		m["wechatOpenId"] = *u.WechatOpenId
	}
	if u.DeptId != nil {
		m["deptId"] = *u.DeptId
	}
	return m
}

// UserWithRolesAndPrivileges 在脱敏用户基础上附加 roles（含 privileges）与 dept。
func UserWithRolesAndPrivileges(u *ent.User, roles []repo.RoleEntity, dept *ent.Dept) map[string]interface{} {
	m := SanitizeUser(u)
	if len(roles) > 0 {
		roleMaps := make([]map[string]interface{}, 0, len(roles))
		for _, r := range roles {
			rm := map[string]interface{}{
				"id":       r.ID,
				"roleName": r.RoleName,
				"roleDesc": r.RoleDesc,
			}
			if len(r.Privileges) > 0 {
				privs := make([]map[string]interface{}, 0, len(r.Privileges))
				for _, p := range r.Privileges {
					privs = append(privs, map[string]interface{}{
						"id":               p.ID,
						"privilegeName":    p.PrivilegeName,
						"privilegeCode":    p.PrivilegeCode,
						"privilegePage":    p.PrivilegePage,
						"pathPattern":      p.PathPattern,
						"httpMethod":       p.HTTPMethod,
						"isPublic":         p.IsPublic,
						"requireOwnership": p.RequireOwnership,
					})
				}
				rm["privileges"] = privs
			}
			roleMaps = append(roleMaps, rm)
		}
		m["roles"] = roleMaps
	}
	if dept != nil {
		m["dept"] = map[string]interface{}{
			"id":       dept.ID,
			"deptName": dept.DeptName,
		}
	}
	return m
}

// UserWithRoles 在脱敏用户基础上附加 roles、dept，供登录/info 接口使用。
func UserWithRoles(u *ent.User, roles []repo.RoleEntity, dept *ent.Dept) map[string]interface{} {
	m := SanitizeUser(u)
	if len(roles) > 0 {
		m["roles"] = rolesToMaps(roles)
	}
	if dept != nil {
		m["dept"] = map[string]interface{}{
			"id":       dept.ID,
			"deptName": dept.DeptName,
		}
	}
	return m
}

// RolePayloadsForJWT 将角色实体转为 JWT payload 中的 role 数组。
func RolePayloadsForJWT(roles []repo.RoleEntity) []JWTRolePayload {
	out := make([]JWTRolePayload, 0, len(roles))
	for _, r := range roles {
		out = append(out, JWTRolePayload{
			ID:       r.ID,
			RoleName: r.RoleName,
			RoleDesc: r.RoleDesc,
		})
	}
	return out
}

func rolesToMaps(roles []repo.RoleEntity) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(roles))
	for _, r := range roles {
		out = append(out, map[string]interface{}{
			"id":       r.ID,
			"roleName": r.RoleName,
			"roleDesc": r.RoleDesc,
		})
	}
	return out
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format("2006-01-02T15:04:05.000Z")
}
