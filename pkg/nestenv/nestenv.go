// Package nestenv 解析 NestJS blog-server deploy/pm2/env.production（key = value）格式。
package nestenv

import (
	"os"
	"regexp"
	"strconv"
	"strings"
)

// ParseFile 读取 Nest env 文件为 key→value 映射。
func ParseFile(path string) (map[string]string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	out := make(map[string]string)
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.Index(line, "=")
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := trimValue(line[idx+1:])
		if key != "" {
			out[key] = val
		}
	}
	return out, nil
}

var inlineComment = regexp.MustCompile(`\s+#`)

func trimValue(s string) string {
	s = strings.TrimSpace(s)
	if i := inlineComment.FindStringIndex(s); i != nil {
		s = strings.TrimSpace(s[:i[0]])
	}
	if len(s) >= 2 {
		if (s[0] == '\'' && s[len(s)-1] == '\'') || (s[0] == '"' && s[len(s)-1] == '"') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

// Get 读取键值。
func Get(m map[string]string, key string) string {
	return strings.TrimSpace(m[key])
}

// SplitCSV 解析英文逗号分隔列表。
func SplitCSV(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// RedisAddr 组装 redis addr。
func RedisAddr(m map[string]string) string {
	host := Get(m, "redis_host")
	if host == "" {
		host = "127.0.0.1"
	}
	port := Get(m, "redis_port")
	if port == "" {
		port = "6379"
	}
	return host + ":" + port
}

// RedisDB 解析 redis_db。
func RedisDB(m map[string]string) int {
	n, _ := strconv.Atoi(Get(m, "redis_db"))
	return n
}

// MySQLBlock 生成 Go mysql yaml 段。
func MySQLBlock(m map[string]string) map[string]any {
	port, _ := strconv.Atoi(Get(m, "db_port"))
	if port == 0 {
		port = 3306
	}
	db := Get(m, "db_database")
	if db == "" {
		db = "myblog"
	}
	return map[string]any{
		"host":         Get(m, "db_host"),
		"port":         port,
		"user":         Get(m, "db_username"),
		"password":     Get(m, "db_password"),
		"database":     db,
		"table_prefix": "x_",
	}
}

// JWTBlock 生成 jwt yaml 段。
func JWTBlock(m map[string]string) map[string]any {
	return map[string]any{
		"secret":      Get(m, "auth_jwtSecret"),
		"legacy_ttl":  "8h",
		"access_ttl":  "30m",
		"refresh_ttl": "168h",
	}
}

func observabilityBlock(serviceName string) map[string]any {
	return map[string]any{
		"enable_metrics": true,
		"enable_pprof":   false,
		"service_name":   serviceName,
	}
}

func fallback(v, def string) string {
	if strings.TrimSpace(v) == "" {
		return def
	}
	return v
}

func envBool(m map[string]string, key string, def bool) bool {
	v := strings.ToLower(strings.TrimSpace(Get(m, key)))
	switch v {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return def
	}
}

func envInt(m map[string]string, key string, def int) int {
	v := strings.TrimSpace(Get(m, key))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

// RagBlock 从 Nest rag_* 环境变量生成 Go rag yaml 段（对齐 blog-server loadRagConfig）。
func RagBlock(m map[string]string) map[string]any {
	embedKey := Get(m, "rag_embedding_api_key")
	embedURL := fallback(Get(m, "rag_embedding_api_base_url"), "https://api.siliconflow.cn/v1")
	mode := "local"
	if embedKey != "" || Get(m, "rag_embedding_api_base_url") != "" {
		mode = "remote"
	}
	return map[string]any{
		"enabled":              envBool(m, "rag_enabled", false),
		"daily_quota":            envInt(m, "rag_daily_query_limit", 20),
		"top_k":                  envInt(m, "rag_top_k", 6),
		"allow_local_fallback":   envBool(m, "rag_allow_local_fallback", true),
		"embedding": map[string]any{
			"mode":        mode,
			"remote_url":  embedURL,
			"api_key":     embedKey,
			"model":       fallback(Get(m, "rag_embedding_model"), "BAAI/bge-large-zh-v1.5"),
		},
		"llm": map[string]any{
			"base_url": fallback(Get(m, "rag_api_base_url"), "https://api.deepseek.com/v1"),
			"api_key":  Get(m, "rag_api_key"),
			"model":    fallback(Get(m, "rag_chat_model"), "deepseek-chat"),
		},
		"chunk": map[string]any{
			"size":    600,
			"overlap": 120,
		},
	}
}

// GatewayYAML 从 Nest env 生成 gateway 配置文档。
func GatewayYAML(m map[string]string) map[string]any {
	return map[string]any{
		"app": map[string]any{
			"name":         "api-gateway",
			"env":          "production",
			"service_mode": "gateway",
			"api_prefix":   "/api/v1",
			"blog_home":    Get(m, "app_blogHome"),
		},
		"http": map[string]any{
			"addr":         ":8000",
			"cors_origins": SplitCSV(Get(m, "serve_corsOrigins")),
		},
		"proxy": map[string]any{
			"user_url": "http://127.0.0.1:5002",
			"blog_url": "http://127.0.0.1:5001",
			"rpg_url":  "http://127.0.0.1:5003",
		},
		"jwt":           JWTBlock(m),
		"observability": observabilityBlock("api-gateway"),
	}
}

// UserYAML 从 Nest env 生成 user 配置文档。
func UserYAML(m map[string]string) map[string]any {
	port, _ := strconv.Atoi(Get(m, "app_emailPort"))
	return map[string]any{
		"app": map[string]any{
			"name":         "user-service",
			"env":          "production",
			"service_mode": "user",
			"api_prefix":   "/api/v1",
			"blog_home":    Get(m, "app_blogHome"),
		},
		"http": map[string]any{"addr": ":5002"},
		"grpc": map[string]any{"addr": ":50052"},
		"mysql": MySQLBlock(m),
		"redis": map[string]any{
			"addr": RedisAddr(m),
			"db":   RedisDB(m),
		},
		"jwt":    JWTBlock(m),
		"crypto": map[string]any{"rsa_private_key": ""},
		"oauth": map[string]any{
			"github_client_id":     Get(m, "app_githubClientId"),
			"github_client_secret": Get(m, "app_githubClientSecret"),
			"github_callback_url":  Get(m, "app_githubCallbackUrl"),
		},
		"mail": map[string]any{
			"host": Get(m, "app_emailHost"),
			"port": port,
			"user": Get(m, "app_emailUser"),
			"pass": Get(m, "app_emailPass"),
		},
		"wechat": map[string]any{
			"app_id": Get(m, "pay_wechatAppId"),
			"secret": Get(m, "pay_wechatSecret"),
		},
		"observability": observabilityBlock("user-service"),
	}
}

// BlogYAML 从 Nest env 生成 blog 配置文档。
func BlogYAML(m map[string]string) map[string]any {
	upload := Get(m, "file_filePath")
	if upload == "" {
		upload = "./public/uploads/"
	}
	return map[string]any{
		"app": map[string]any{
			"name":                 "blog-service",
			"env":                  "production",
			"service_mode":         "blog",
			"api_prefix":           "/api/v1",
			"blog_home":            Get(m, "app_blogHome"),
			"notify_email":         Get(m, "app_notifyEmail"),
			"tongji_refresh_token": Get(m, "app_tongjiRefreshToken"),
			"tongji_client_id":     Get(m, "app_tongjiClientId"),
			"tongji_client_secret": Get(m, "app_tongjiClientSecret"),
		},
		"http": map[string]any{"addr": ":5001"},
		"grpc": map[string]any{"user_addr": "127.0.0.1:50052"},
		"mysql": MySQLBlock(m),
		"redis": map[string]any{
			"addr": RedisAddr(m),
			"db":   RedisDB(m),
		},
		"jwt": JWTBlock(m),
		"storage": map[string]any{
			"upload_path":   upload,
			"public_prefix": "/static/",
		},
		"rag":           RagBlock(m),
		"observability": observabilityBlock("blog-service"),
	}
}

// RpgYAML 从 Nest env 生成 rpg 配置文档。
func RpgYAML(m map[string]string) map[string]any {
	upload := Get(m, "file_filePath")
	if upload == "" {
		upload = "./public/uploads/"
	}
	return map[string]any{
		"app": map[string]any{
			"name":         "rpg-service",
			"env":          "production",
			"service_mode": "rpg",
			"api_prefix":   "/api/v1",
			"blog_home":    Get(m, "app_blogHome"),
		},
		"http": map[string]any{"addr": ":5003"},
		"grpc": map[string]any{"user_addr": "127.0.0.1:50052"},
		"mysql": MySQLBlock(m),
		"redis": map[string]any{
			"addr": RedisAddr(m),
			"db":   RedisDB(m),
		},
		"jwt": JWTBlock(m),
		"storage": map[string]any{
			"upload_path":   upload,
			"public_prefix": "/static/",
		},
		"pay": map[string]any{
			"alipay_app_id":              Get(m, "pay_alipayAppId"),
			"alipay_private_key":         Get(m, "pay_alipayPrivateKey"),
			"alipay_public_key":          Get(m, "pay_alipayPublicKey"),
			"alipay_gateway":             fallback(Get(m, "pay_alipayGateway"), "https://openapi.alipay.com/gateway.do"),
			"alipay_notify_url":          Get(m, "pay_alipayNotifyUrl"),
			"alipay_return_url":          Get(m, "pay_alipayReturnUrl"),
			"alipay_mini_cashier_page":   fallback(Get(m, "pay_alipayMiniCashierPage"), "packageB/pages/business/pay/all-pay/all-pay"),
			"sandbox":                    false,
			"use_legacy_sandbox_gateway": false,
			"wechat_app_id":              Get(m, "pay_wechatAppId"),
			"wechat_secret":              Get(m, "pay_wechatSecret"),
		},
		"observability": observabilityBlock("rpg-service"),
	}
}
