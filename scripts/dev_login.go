//go:debug rsa1024min=0

// 本地开发登录：获取超级管理员 JWT，供 curl/联调/测试使用。
//
//	go run scripts/dev_login.go
//	go run scripts/dev_login.go --token-only
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/Jiang-Xia/blog-server-go/internal/devlogin"
	"github.com/Jiang-Xia/blog-server-go/pkg/config"
)

func main() {
	tokenOnly := flag.Bool("token-only", false, "仅输出 accessToken")
	asJSON := flag.Bool("json", false, "输出 JSON")
	flag.Parse()

	cfg, err := config.MustLoad("")
	if err != nil {
		fail(err)
	}

	username := envOr("TEST_USERNAME", devlogin.DefaultUsername)
	password := envOr("TEST_PASSWORD", devlogin.DefaultPassword)
	base := strings.TrimRight(envOr("DEV_LOGIN_BASE", "http://127.0.0.1:8000"), "/")
	apiPrefix := strings.Trim(cfg.App.APIPrefix, "/")
	if apiPrefix == "" {
		apiPrefix = "api/v1"
	}

	tokens, err := devlogin.Login(context.Background(), cfg, base+"/"+apiPrefix, username, password)
	if err != nil {
		fail(err)
	}

	switch {
	case *asJSON:
		b, _ := json.MarshalIndent(tokens, "", "  ")
		fmt.Println(string(b))
	case *tokenOnly:
		fmt.Println(tokens.AccessToken)
	default:
		fmt.Printf("username: %s\n", username)
		fmt.Printf("accessToken: %s\n", tokens.AccessToken)
		fmt.Printf("refreshToken: %s\n", tokens.RefreshToken)
		fmt.Printf("\n# curl 示例\n")
		fmt.Printf("curl.exe -H \"Authorization: Bearer %s\" %s/%s/user/info\n",
			tokens.AccessToken, base, apiPrefix)
	}
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, "dev_login: %v\n", err)
	os.Exit(1)
}
