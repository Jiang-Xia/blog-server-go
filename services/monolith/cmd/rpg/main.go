//go:debug rsa1024min=0

// rpg-service 入口。
package main

import (
	"log"
	"os"

	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/app"
)

func main() {
	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = "configs/rpg.yaml"
	}
	application, err := app.InitializeApp(cfgPath)
	if err != nil {
		log.Fatalf("initialize rpg-service: %v", err)
	}
	if err := application.Run(); err != nil {
		log.Fatalf("run rpg-service: %v", err)
	}
}
