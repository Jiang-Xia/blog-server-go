// Package middleware 操作日志中间件，对齐 Nest OperationLogInterceptor。
package middleware

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/Jiang-Xia/blog-server-go/pkg/ctxutil"
	"github.com/Jiang-Xia/blog-server-go/services/user/internal/operationlog"
	"github.com/Jiang-Xia/blog-server-go/services/user/internal/user/auth"
	"github.com/cloudwego/hertz/pkg/app"
	"go.uber.org/zap"
)

var (
	loggedMethods = map[string]bool{
		"POST": true, "PUT": true, "PATCH": true, "DELETE": true,
	}
	skipPathPatterns = []*regexp.Regexp{
		regexp.MustCompile(`/article/list$`),
		regexp.MustCompile(`/article/views$`),
		regexp.MustCompile(`/article/likes$`),
		regexp.MustCompile(`/like$`),
		regexp.MustCompile(`/collect(/|$)`),
		regexp.MustCompile(`/comment/create$`),
		regexp.MustCompile(`/reply/create$`),
		regexp.MustCompile(`/msgboard$`),
		regexp.MustCompile(`/user/login$`),
		regexp.MustCompile(`/user/register$`),
		regexp.MustCompile(`/user/email/`),
		regexp.MustCompile(`/user/authCode$`),
		regexp.MustCompile(`/captcha(/|$)`),
		regexp.MustCompile(`/operation-log(/|$)`),
		regexp.MustCompile(`/pub/`),
		regexp.MustCompile(`/pay/`),
		regexp.MustCompile(`/file/`),
		regexp.MustCompile(`/resources/uploadFile$`),
		regexp.MustCompile(`/resources/upload-media(/|$)`),
	}
	sensitiveBodyFields = []string{"password", "oldPassword", "newPassword", "token", "secret", "salt"}
)

// OperationLogDeps 操作日志中间件依赖。
type OperationLogDeps struct {
	Svc *operationlog.Service
	JWT *auth.JWTService
	Log *zap.Logger
}

// OperationLog 记录写操作到 operation_log 表（异步 fire-and-forget）。
func OperationLog(deps OperationLogDeps) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		method := strings.ToUpper(string(c.Method()))
		path := string(c.Path())

		if !loggedMethods[method] || shouldSkipOpLog(path) {
			c.Next(ctx)
			return
		}

		bodyCopy := copyRequestBody(c)
		c.Next(ctx)

		uid, username := extractOpLogUser(ctx, c, deps.JWT)
		ip := clientIP(c)
		module := extractModule(path)
		action := extractAction(method)
		desc := buildOpDescription(method, module, path)
		statusCode := c.Response.StatusCode()
		if statusCode == 0 {
			statusCode = 200
		}

		params := operationlog.CreateParams{
			UserID:      uid,
			Username:    username,
			Module:      module,
			Action:      action,
			Method:      method,
			Path:        path,
			Description: desc,
			IP:          ip,
			RequestBody: sanitizeBody(bodyCopy),
			StatusCode:  statusCode,
		}
		go func() {
			if err := deps.Svc.Create(context.Background(), params); err != nil {
				deps.Log.Warn("操作日志写入失败", zap.Error(err), zap.String("path", path))
			}
		}()
	}
}

func shouldSkipOpLog(path string) bool {
	clean := strings.SplitN(path, "?", 2)[0]
	for _, re := range skipPathPatterns {
		if re.MatchString(clean) {
			return true
		}
	}
	return false
}

func extractModule(path string) string {
	cleaned := regexp.MustCompile(`^/api/v\d+`).ReplaceAllString(path, "")
	cleaned = strings.TrimPrefix(cleaned, "/")
	parts := strings.Split(cleaned, "/")
	for _, p := range parts {
		if p != "" {
			return p
		}
	}
	return "unknown"
}

func extractAction(method string) string {
	switch method {
	case "POST":
		return "create"
	case "PUT", "PATCH":
		return "update"
	case "DELETE":
		return "delete"
	default:
		return "other"
	}
}

func buildOpDescription(method, module, path string) string {
	actionText := map[string]string{
		"POST": "新增", "PUT": "修改", "PATCH": "修改", "DELETE": "删除",
	}[method]
	if actionText == "" {
		actionText = "操作"
	}
	return actionText + " " + module + " - " + method + " " + path
}

func extractOpLogUser(ctx context.Context, c *app.RequestContext, jwt *auth.JWTService) (int, string) {
	if uid := ctxutil.UserID(ctx); uid != 0 {
		name := lookupUsernameFromJWT(c, jwt)
		if name == "" {
			name = "unknown"
		}
		return uid, name
	}
	if jwt == nil {
		return 0, "anonymous"
	}
	authz := strings.TrimSpace(string(c.GetHeader("Authorization")))
	if authz == "" {
		return 0, "anonymous"
	}
	token := strings.TrimPrefix(authz, "Bearer ")
	claims, err := jwt.Verify(strings.TrimSpace(token))
	if err != nil || claims == nil {
		return 0, "anonymous"
	}
	name := claims.Nickname
	if name == "" {
		name = claims.Username
	}
	if name == "" {
		name = "unknown"
	}
	return claims.ID, name
}

func lookupUsernameFromJWT(c *app.RequestContext, jwt *auth.JWTService) string {
	if jwt == nil {
		return ""
	}
	authz := strings.TrimSpace(string(c.GetHeader("Authorization")))
	if authz == "" {
		return ""
	}
	token := strings.TrimPrefix(authz, "Bearer ")
	claims, err := jwt.Verify(strings.TrimSpace(token))
	if err != nil || claims == nil {
		return ""
	}
	if claims.Nickname != "" {
		return claims.Nickname
	}
	return claims.Username
}

func clientIP(c *app.RequestContext) string {
	if xff := string(c.GetHeader("X-Forwarded-For")); xff != "" {
		return strings.TrimSpace(strings.Split(xff, ",")[0])
	}
	if xri := string(c.GetHeader("X-Real-Ip")); xri != "" {
		return strings.TrimSpace(xri)
	}
	return c.ClientIP()
}

func copyRequestBody(c *app.RequestContext) []byte {
	body := c.Request.Body()
	if len(body) == 0 {
		return nil
	}
	dup := make([]byte, len(body))
	copy(dup, body)
	return dup
}

func sanitizeBody(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return ""
	}
	maskSensitiveFields(data)
	raw, err := json.Marshal(data)
	if err != nil {
		return "[无法序列化]"
	}
	s := string(raw)
	if len(s) > 2000 {
		return s[:2000] + "...(truncated)"
	}
	return s
}

func maskSensitiveFields(obj map[string]interface{}) {
	for k, v := range obj {
		for _, field := range sensitiveBodyFields {
			if k == field {
				obj[k] = "***"
			}
		}
		if nested, ok := v.(map[string]interface{}); ok {
			maskSensitiveFields(nested)
		}
	}
}
