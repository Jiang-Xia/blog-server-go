// Package data 封装 user-service Ent 与 Redis 客户端。
package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/services/user/ent"
	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/rueidis"
	"go.uber.org/zap"
)

// NewEntClient 创建 Ent MySQL 客户端并验证连通性。
func NewEntClient(cfg *config.Config, log *zap.Logger) (*ent.Client, error) {
	dsn := cfg.MySQL.FormatDSN()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sql db: %w", err)
	}
	db.SetMaxOpenConns(15)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	drv := entsql.OpenDB(dialect.MySQL, db)
	client := ent.NewClient(ent.Driver(drv))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := client.User.Query().Limit(1).All(ctx); err != nil {
		client.Close()
		return nil, fmt.Errorf("ping user table: %w", err)
	}
	log.Info("ent client connected")
	return client, nil
}

// NewRedisClient 创建 rueidis 客户端并 PING。
func NewRedisClient(cfg *config.Config, log *zap.Logger) (rueidis.Client, error) {
	client, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress:  []string{cfg.Redis.Addr},
		SelectDB:     cfg.Redis.DB,
		DisableCache: true,
	})
	if err != nil {
		return nil, fmt.Errorf("new redis client: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Do(ctx, client.B().Ping().Build()).Error(); err != nil {
		client.Close()
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	log.Info("redis client connected", zap.String("addr", cfg.Redis.Addr), zap.Int("db", cfg.Redis.DB))
	return client, nil
}

// CloseEnt 关闭 Ent 客户端。
func CloseEnt(client *ent.Client) {
	if client != nil {
		client.Close()
	}
}

// CloseRedis 关闭 Redis 客户端。
func CloseRedis(client rueidis.Client) {
	if client != nil {
		client.Close()
	}
}
