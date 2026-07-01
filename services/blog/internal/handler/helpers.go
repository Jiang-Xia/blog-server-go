// helpers handler 层共用请求解析与响应辅助。
package handler

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Jiang-Xia/blog-server-go/pkg/response"
	"github.com/cloudwego/hertz/pkg/app"
)

func strField(m map[string]interface{}, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	return fmt.Sprint(v)
}

func intField(m map[string]interface{}, key string) int {
	v, ok := m[key]
	if !ok {
		return 0
	}
	i, _ := toInt(v)
	return i
}

func toInt(v interface{}) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int64:
		return int(n), true
	case float64:
		return int(n), true
	case string:
		i, err := strconv.Atoi(n)
		return i, err == nil
	default:
		return 0, false
	}
}

func intFieldDefault(m map[string]interface{}, key string, def int) int {
	if m[key] == nil {
		return def
	}
	return intField(m, key)
}

func handleAdminResult(ctx context.Context, c *app.RequestContext, data interface{}, err error) {
	if err != nil {
		response.FromError(ctx, c, err)
		return
	}
	response.Success(ctx, c, data)
}
