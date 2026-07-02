package nestenv

import (
	"strconv"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
)

// MySQLConfig 从 Nest env 映射为 Go MySQL 配置（table_prefix 固定 x_）。
func MySQLConfig(m map[string]string) config.MySQLConfig {
	port, _ := strconv.Atoi(Get(m, "db_port"))
	if port == 0 {
		port = 3306
	}
	db := Get(m, "db_database")
	if db == "" {
		db = "myblog"
	}
	return config.MySQLConfig{
		Host:     Get(m, "db_host"),
		Port:     port,
		User:     Get(m, "db_username"),
		Password: Get(m, "db_password"),
		Database: db,
		TablePrefix: "x_",
		SchemaSourceDatabase: "myblog",
	}
}

// RedisConfig 从 Nest env 映射 Redis 配置。
func RedisConfig(m map[string]string) config.RedisConfig {
	return config.RedisConfig{
		Addr: RedisAddr(m),
		DB:   RedisDB(m),
	}
}

// ConfigFromFile 读取 env.production 并组装 bootstrap/sync 用的最小 Config。
func ConfigFromFile(path string) (*config.Config, error) {
	m, err := ParseFile(path)
	if err != nil {
		return nil, err
	}
	return &config.Config{
		MySQL: MySQLConfig(m),
		Redis: RedisConfig(m),
	}, nil
}
