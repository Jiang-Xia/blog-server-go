// Package main 的 CI 环境变量与 DSN 辅助。
package main

import (
	"fmt"
	"os"
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
	host := env("CI_MYSQL_HOST", "127.0.0.1")
	port := envInt("CI_MYSQL_PORT", 3306)
	user := env("CI_MYSQL_USER", defaultMySQLUser)
	pass := env("CI_MYSQL_PASSWORD", defaultMySQLPass)
	db := env("CI_MYSQL_DATABASE", defaultMySQLDB)
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=Local&charset=utf8mb4",
		user, pass, host, port, db)
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
