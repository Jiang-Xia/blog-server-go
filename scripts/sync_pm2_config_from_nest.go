// sync_pm2_config_from_nest 从 deploy/pm2/env.production 生成 PM2 四服务 configs/*.yaml。
// 用法：go run scripts/sync_pm2_config_from_nest.go --env deploy/pm2/env.production
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Jiang-Xia/blog-server-go/pkg/nestenv"
	"gopkg.in/yaml.v3"
)

func main() {
	envPath := flag.String("env", "deploy/pm2/env.production", "生产 env 文件路径（与 blog-server 同格式）")
	outDir := flag.String("out", "deploy/pm2/configs", "输出目录")
	flag.Parse()

	env, err := nestenv.ParseFile(*envPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse env: %v\n", err)
		os.Exit(1)
	}

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir out: %v\n", err)
		os.Exit(1)
	}

	files := map[string]map[string]any{
		"gateway.yaml": nestenv.GatewayYAML(env),
		"user.yaml":    nestenv.UserYAML(env),
		"blog.yaml":    nestenv.BlogYAML(env),
		"rpg.yaml":     nestenv.RpgYAML(env),
	}

	for name, doc := range files {
		path := filepath.Join(*outDir, name)
		data, err := yaml.Marshal(doc)
		if err != nil {
			fmt.Fprintf(os.Stderr, "marshal %s: %v\n", name, err)
			os.Exit(1)
		}
		header := fmt.Sprintf("# 由 scripts/sync_pm2_config_from_nest.go 从 %s 生成，勿手改；改 env.production 后重新 deploy\n", *envPath)
		if err := os.WriteFile(path, append([]byte(header), data...), 0o600); err != nil {
			fmt.Fprintf(os.Stderr, "write %s: %v\n", path, err)
			os.Exit(1)
		}
		fmt.Printf("wrote %s\n", path)
	}
}
