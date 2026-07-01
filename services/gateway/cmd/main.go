//go:debug rsa1024min=0

// gateway 入口：REST BFF + 反向代理，默认 configs/gateway.yaml、:8000。
package main

import (
	"log"
	"os"

	gwapp "github.com/Jiang-Xia/blog-server-go/services/gateway/internal/app"
)

func main() {
	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = "configs/gateway.yaml"
	}
	if err := gwapp.Run(cfgPath); err != nil {
		log.Fatalf("gateway: %v", err)
	}
}
