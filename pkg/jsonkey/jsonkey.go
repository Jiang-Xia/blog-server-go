// jsonkey API JSON 字段名工具，对齐 Nest TypeORM 小驼峰。
package jsonkey

import "strings"

// SnakeToCamel 将 snake_case 转为小驼峰（article_id → articleId）。
func SnakeToCamel(s string) string {
	if !strings.Contains(s, "_") {
		return s
	}
	parts := strings.Split(s, "_")
	for i := 1; i < len(parts); i++ {
		p := parts[i]
		if p == "" {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, "")
}
