//go:debug rsa1024min=0

// 模块化单体入口：wire 装配后启动 Hertz，监听 CONFIG_PATH 或默认 configs/monolith.yaml。
package main

import (
	"log"
	"os"

	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/app"
)

func main() {
	cfgPath := os.Getenv("CONFIG_PATH")
	application, err := app.InitializeApp(cfgPath)
	if err != nil {
		log.Fatalf("initialize app: %v", err)
	}
	if err := application.Run(); err != nil {
		log.Fatalf("run app: %v", err)
	}
}
