// Package main 的 CI 环境变量与 DSN 辅助。
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
)

const (
	defaultJWTSecret = "ci-integration-test-secret"
	defaultMySQLUser = "root"
	defaultMySQLPass = "testpass"
	defaultMySQLDB   = "x_my_blog"
	defaultRedisDB   = 2
)

func env(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func mysqlDSN() string {
	if env("CI_USE_LOCAL_CONFIG", "") == "1" {
		if dsn, ok := mysqlDSNFromYAML(); ok {
			return dsn
		}
	}
	host := env("CI_MYSQL_HOST", "127.0.0.1")
	port := envInt("CI_MYSQL_PORT", 3306)
	user := env("CI_MYSQL_USER", defaultMySQLUser)
	pass := env("CI_MYSQL_PASSWORD", defaultMySQLPass)
	db := env("CI_MYSQL_DATABASE", defaultMySQLDB)
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=Local&charset=utf8mb4",
		user, pass, host, port, db)
}

// mysqlDSNFromYAML 读取 configs/user.yaml 的 MySQL 段（SkipDocker 本地测试不覆盖 yaml 时用）。
func mysqlDSNFromYAML() (string, bool) {
	root := env("CI_PROJECT_ROOT", ".")
	cfgPath := env("CONFIG_PATH", "configs/user.yaml")
	if !filepath.IsAbs(cfgPath) {
		cfgPath = filepath.Join(root, cfgPath)
	}
	cfg, err := config.MustLoad(cfgPath)
	if err != nil {
		return "", false
	}
	m := cfg.MySQL
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=Local&charset=utf8mb4",
		m.User, m.Password, m.Host, m.Port, m.Database), true
}

func mysqlDSNNoDB() string {
	host := env("CI_MYSQL_HOST", "127.0.0.1")
	port := envInt("CI_MYSQL_PORT", 3306)
	user := env("CI_MYSQL_USER", defaultMySQLUser)
	pass := env("CI_MYSQL_PASSWORD", defaultMySQLPass)
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/?parseTime=true&loc=Local&charset=utf8mb4",
		user, pass, host, port)
}

func redisAddr() string {
	return env("CI_REDIS_ADDR", "127.0.0.1:6379")
}

func jwtSecret() string {
	return env("CI_JWT_SECRET", defaultJWTSecret)
}

func testBaseURL() string {
	return strings.TrimRight(env("TEST_BASE", "http://127.0.0.1:8000"), "/")
}

func loadCIConfig() (*config.Config, error) {
	path := env("CONFIG_PATH", "configs/user.yaml")
	return config.MustLoad(path)
}
