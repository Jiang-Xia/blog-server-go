// 生产库 x_my_blog 初始化：从 deploy/pm2/env.production 读连接，从 myblog 克隆 x_* 表结构+数据。
// 在服务器本机执行（jxblog 仅 localhost）。
//
//	# 1) root 建库授权
//	sudo mysql -u root -p < deploy/sql/prod/001_create_x_my_blog.sql
//
//	# 2) 结构 + 数据（保留 myblog 不动）
//	go run scripts/bootstrap_prod_db.go --env deploy/pm2/env.production
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
)

func main() {
	envPath := flag.String("env", "deploy/pm2/env.production", "生产 env 文件")
	skipRedis := flag.Bool("skip-redis", true, "跳过 Redis FLUSHDB（生产默认 true）")
	sourceDB := flag.String("source", "myblog", "Nest 源库名")
	flag.Parse()

	if _, err := os.Stat(*envPath); err != nil {
		fmt.Fprintf(os.Stderr, "missing env file %s: %v\n", *envPath, err)
		os.Exit(1)
	}

	fmt.Println("==> [1/2] bootstrap table structure (myblog -> x_my_blog, prefix x_)")
	if err := runGo("scripts/bootstrap_x_my_blog.go", *envPath, *sourceDB); err != nil {
		fail(err)
	}

	fmt.Println("==> [2/2] sync data from myblog")
	if err := runGo("scripts/sync_data_x_my_blog.go", *envPath, *sourceDB, *skipRedis); err != nil {
		fail(err)
	}

	fmt.Println("==> bootstrap_prod_db done")
	fmt.Println("    verify: mysql -u jxblog -p -e \"SHOW TABLES FROM x_my_blog LIKE 'x_%'\"")
}

func runGo(script, envPath, sourceDB string, skipRedis ...bool) error {
	args := []string{"run", script, "--env", envPath, "--source", sourceDB}
	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	if len(skipRedis) > 0 && skipRedis[0] {
		cmd.Env = append(cmd.Env, "SYNC_SKIP_REDIS=1")
	}
	return cmd.Run()
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, "bootstrap_prod_db: %v\n", err)
	os.Exit(1)
}
