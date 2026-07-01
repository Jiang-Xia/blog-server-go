// Package data 封装 rpg-service Ent、Redis 与共享 SQL 连接。
package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/ent"
	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/rueidis"
	"go.uber.org/zap"
)

// Data 聚合基础设施客户端。
type Data struct {
	Ent   *ent.Client
	SQL   *sql.DB
	Redis rueidis.Client
}

// NewData 创建 Ent、SQL、Redis 并验证连通性。
func NewData(cfg *config.Config, log *zap.Logger) (*Data, error) {
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
	if _, err := client.Rpg.Query().Limit(1).All(ctx); err != nil {
		client.Close()
		db.Close()
		return nil, fmt.Errorf("ping rpg table: %w", err)
	}

	rds, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress:  []string{cfg.Redis.Addr},
		SelectDB:     cfg.Redis.DB,
		DisableCache: true,
	})
	if err != nil {
		client.Close()
		db.Close()
		return nil, fmt.Errorf("new redis client: %w", err)
	}
	ctx2, cancel2 := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel2()
	if err := rds.Do(ctx2, rds.B().Ping().Build()).Error(); err != nil {
		client.Close()
		db.Close()
		rds.Close()
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	log.Info("rpg-service data connected")
	return &Data{Ent: client, SQL: db, Redis: rds}, nil
}

// Close 关闭全部连接。
func (d *Data) Close() {
	if d.Ent != nil {
		d.Ent.Close()
	}
	if d.SQL != nil {
		d.SQL.Close()
	}
	if d.Redis != nil {
		d.Redis.Close()
	}
}
