//go:debug rsa1024min=0

// blog-service 入口。
package main

import (
	"log"
	"os"

	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/app"
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
