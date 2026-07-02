// 为 CI / 本地测试流水线生成 configs/*.yaml（覆盖写入，勿用于生产）。
//
//	go run scripts/ci/prepare_config.go
package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	root := env("CI_PROJECT_ROOT", ".")
	secret := jwtSecret()
	mysqlHost := env("CI_MYSQL_HOST", "127.0.0.1")
	mysqlPort := envInt("CI_MYSQL_PORT", 3306)
	mysqlUser := env("CI_MYSQL_USER", defaultMySQLUser)
	mysqlPass := env("CI_MYSQL_PASSWORD", defaultMySQLPass)
	mysqlDB := env("CI_MYSQL_DATABASE", defaultMySQLDB)
	redisAddr := redisAddr()
	redisDB := envInt("CI_REDIS_DB", defaultRedisDB)

	files := map[string]string{
		"configs/user.yaml": userYAML(mysqlHost, mysqlPort, mysqlUser, mysqlPass, mysqlDB, redisAddr, redisDB, secret),
		"configs/blog.yaml": serviceYAML("blog-service", "blog", ":5001", ":50051", mysqlHost, mysqlPort, mysqlUser, mysqlPass, mysqlDB, redisAddr, redisDB, secret, ":6061"),
		"configs/rpg.yaml":  serviceYAML("rpg-service", "rpg", ":5003", ":50053", mysqlHost, mysqlPort, mysqlUser, mysqlPass, mysqlDB, redisAddr, redisDB, secret, ":6063"),
		"configs/gateway.yaml": gatewayYAML(mysqlHost, mysqlPort, mysqlUser, mysqlPass, mysqlDB, redisAddr, redisDB, secret),
		"configs/monolith.yaml": userYAML(mysqlHost, mysqlPort, mysqlUser, mysqlPass, mysqlDB, redisAddr, redisDB, secret),
	}

	for rel, content := range files {
		path := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			fail(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			fail(err)
		}
		fmt.Println("wrote", rel)
	}
}

func userYAML(host string, port int, user, pass, db, redis string, redisDB int, secret string) string {
	return fmt.Sprintf(`app:
  name: user-service
  env: development
  service_mode: user
  api_prefix: /api/v1

http:
  addr: ":5002"

grpc:
  addr: ":50052"

mysql:
  host: %s
  port: %d
  user: %s
  password: %s
  database: %s
  table_prefix: x_

redis:
  addr: "%s"
  db: %d

jwt:
  secret: "%s"
  legacy_ttl: 8h
  access_ttl: 30m
  refresh_ttl: 168h

crypto:
  rsa_private_key: ""

observability:
  enable_metrics: false
  enable_pprof: false
  service_name: user-service
`, host, port, user, pass, db, redis, redisDB, secret)
}

func serviceYAML(name, mode, httpAddr, grpcAddr, host string, port int, user, pass, db, redis string, redisDB int, secret, pprof string) string {
	grpcBlock := ""
	if grpcAddr != "" {
		grpcBlock = fmt.Sprintf(`
grpc:
  addr: "%s"
  user_addr: "127.0.0.1:50052"
`, grpcAddr)
	} else if mode == "rpg" {
		grpcBlock = `
grpc:
  addr: ":50053"
  user_addr: "127.0.0.1:50052"
`
	}
	return fmt.Sprintf(`app:
  name: %s
  env: development
  service_mode: %s
  api_prefix: /api/v1

http:
  addr: "%s"
%s
mysql:
  host: %s
  port: %d
  user: %s
  password: %s
  database: %s
  table_prefix: x_

redis:
  addr: "%s"
  db: %d

jwt:
  secret: "%s"
  legacy_ttl: 8h
  access_ttl: 30m
  refresh_ttl: 168h

crypto:
  rsa_private_key: ""

observability:
  enable_metrics: false
  enable_pprof: false
  pprof_addr: "%s"
  service_name: %s
`, name, mode, httpAddr, grpcBlock, host, port, user, pass, db, redis, redisDB, secret, pprof, name)
}

func gatewayYAML(host string, port int, user, pass, db, redis string, redisDB int, secret string) string {
	return fmt.Sprintf(`app:
  name: api-gateway
  env: development
  service_mode: gateway
  api_prefix: /api/v1

http:
  addr: ":8000"

proxy:
  user_url: "http://127.0.0.1:5002"
  blog_url: "http://127.0.0.1:5001"
  rpg_url: "http://127.0.0.1:5003"

grpc:
  user_addr: "127.0.0.1:50052"
  blog_addr: "127.0.0.1:50051"
  rpg_addr: "127.0.0.1:50053"

jwt:
  secret: "%s"
  legacy_ttl: 8h
  access_ttl: 30m
  refresh_ttl: 168h

observability:
  enable_metrics: false
  enable_pprof: false
  service_name: api-gateway
`, secret)
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, "ci: %v\n", err)
	os.Exit(1)
}
