// userinfo 从 user 域批量拉取作者展示字段。
package util

import "github.com/Jiang-Xia/blog-server-go/pkg/usersvc"

// UserInfoMap 将 UserDTO 转为 Nest setUserInfo 风格 map。
func UserInfoMap(u *usersvc.UserDTO, fields ...string) map[string]interface{} {
	if u == nil {
		return map[string]interface{}{}
	}
	all := map[string]interface{}{
		"id":       u.ID,
		"nickname": u.Nickname,
		"username": u.Username,
		"avatar":   u.Avatar,
	}
	if len(fields) == 0 {
		return all
	}
	out := make(map[string]interface{}, len(fields))
	for _, f := range fields {
		if v, ok := all[f]; ok {
			out[f] = v
		}
	}
	return out
}
