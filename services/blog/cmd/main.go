//go:debug rsa1024min=0

// blog-service 入口（Plan 11 物理拆分）。
package main

import (
	"log"
	"os"

	_ "github.com/Jiang-Xia/blog-server-go/services/blog/docs"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/app"
)

func main() {
	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = "configs/blog.yaml"
	}
	application, err := app.InitializeApp(cfgPath)
	if err != nil {
		log.Fatalf("initialize blog-service: %v", err)
	}
	if err := application.Run(); err != nil {
		log.Fatalf("run blog-service: %v", err)
	}
}
