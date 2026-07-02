// 全量从 Nest 源库拷贝数据到目标库 x_* 表；可选 FLUSHDB Redis。
// 用法：go run scripts/sync_data_x_my_blog.go --env deploy/pm2/env.production --source myblog
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/nestenv"
	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/rueidis"
)

func loadSyncConfig() (*config.Config, string, error) {
	envPath := flag.String("env", "", "deploy/pm2/env.production 路径")
	sourceDB := flag.String("source", "", "源库名")
	flag.Parse()

	src := strings.TrimSpace(*sourceDB)
	if src == "" {
		src = strings.TrimSpace(os.Getenv("SYNC_SOURCE_DB"))
	}

	if strings.TrimSpace(*envPath) != "" {
		cfg, err := nestenv.ConfigFromFile(*envPath)
		if err != nil {
			return nil, "", err
		}
		if src == "" {
			src = "myblog"
		}
		return cfg, src, nil
	}

	path := strings.TrimSpace(os.Getenv("CONFIG_PATH"))
	if path == "" {
		path = "configs/monolith.yaml"
	}
	cfg, err := config.MustLoad(path)
	if err != nil {
		return nil, "", err
	}
	if src == "" {
		src = strings.TrimSpace(cfg.MySQL.SchemaSourceDatabase)
	}
	if src == "" {
		src = "myblog"
	}
	return cfg, src, nil
}

func main() {
	cfg, sourceDB, err := loadSyncConfig()
	if err != nil {
		panic(err)
	}
	prefix := cfg.MySQL.TablePrefixOrDefault()
	targetDB := cfg.MySQL.Database
	skipRedis := strings.TrimSpace(os.Getenv("SYNC_SKIP_REDIS")) == "1"

	base := cfg.MySQL
	base.Database = ""
	db, err := sql.Open("mysql", base.FormatDSN())
	if err != nil {
		panic(err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		panic(fmt.Errorf("mysql ping: %w", err))
	}

	tables, err := listTables(db, sourceDB)
	if err != nil {
		panic(err)
	}
	if len(tables) == 0 {
		panic(fmt.Sprintf("source database %s has no tables", sourceDB))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	if _, err := db.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS=0"); err != nil {
		panic(fmt.Errorf("disable fk checks: %w", err))
	}
	defer func() { _, _ = db.Exec("SET FOREIGN_KEY_CHECKS=1") }()

	fmt.Printf("truncating %d tables in %s ...\n", len(tables), targetDB)
	for _, t := range tables {
		if strings.HasPrefix(t, prefix) {
			continue
		}
		dest := prefix + t
		q := fmt.Sprintf("TRUNCATE TABLE `%s`.`%s`", targetDB, dest)
		if _, err := db.ExecContext(ctx, q); err != nil {
			panic(fmt.Errorf("truncate %s: %w", dest, err))
		}
	}

	fmt.Printf("copying data %s -> %s (prefix %q) ...\n", sourceDB, targetDB, prefix)
	var total int64
	for _, t := range tables {
		if strings.HasPrefix(t, prefix) {
			continue
		}
		dest := prefix + t
		q := fmt.Sprintf("INSERT INTO `%s`.`%s` SELECT * FROM `%s`.`%s`", targetDB, dest, sourceDB, t)
		res, err := db.ExecContext(ctx, q)
		if err != nil {
			panic(fmt.Errorf("copy %s -> %s: %w", t, dest, err))
		}
		n, _ := res.RowsAffected()
		total += n
		fmt.Printf("  %s.%s -> %s.%s (%d rows)\n", sourceDB, t, targetDB, dest, n)
	}

	filtered := make([]string, 0, len(tables))
	for _, t := range tables {
		if !strings.HasPrefix(t, prefix) {
			filtered = append(filtered, t)
		}
	}
	if err := verifyCounts(ctx, db, sourceDB, targetDB, prefix, filtered); err != nil {
		panic(err)
	}
	fmt.Printf("mysql sync done: %d tables, %d rows copied (source %s unchanged)\n", len(filtered), total, sourceDB)

	if skipRedis {
		fmt.Println("redis flush skipped (SYNC_SKIP_REDIS=1)")
		return
	}
	if err := flushRedis(cfg); err != nil {
		panic(err)
	}
	fmt.Printf("redis flush done: db %d on %s\n", cfg.Redis.DB, cfg.Redis.Addr)
}

func listTables(db *sql.DB, schema string) ([]string, error) {
	rows, err := db.Query(
		"SELECT TABLE_NAME FROM information_schema.TABLES WHERE TABLE_SCHEMA = ? ORDER BY TABLE_NAME",
		schema,
	)
	if err != nil {
		return nil, fmt.Errorf("list tables from %s: %w", schema, err)
	}
	defer rows.Close()
	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tables = append(tables, name)
	}
	return tables, rows.Err()
}

func verifyCounts(ctx context.Context, db *sql.DB, sourceDB, targetDB, prefix string, tables []string) error {
	for _, t := range tables {
		dest := prefix + t
		var srcCnt, dstCnt int64
		if err := db.QueryRowContext(ctx,
			fmt.Sprintf("SELECT COUNT(*) FROM `%s`.`%s`", sourceDB, t),
		).Scan(&srcCnt); err != nil {
			return fmt.Errorf("count source %s: %w", t, err)
		}
		if err := db.QueryRowContext(ctx,
			fmt.Sprintf("SELECT COUNT(*) FROM `%s`.`%s`", targetDB, dest),
		).Scan(&dstCnt); err != nil {
			return fmt.Errorf("count target %s: %w", dest, err)
		}
		if srcCnt != dstCnt {
			return fmt.Errorf("row count mismatch %s: source=%d target=%d", t, srcCnt, dstCnt)
		}
	}
	return nil
}

func flushRedis(cfg *config.Config) error {
	client, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress:  []string{cfg.Redis.Addr},
		SelectDB:     cfg.Redis.DB,
		DisableCache: true,
	})
	if err != nil {
		return fmt.Errorf("new redis client: %w", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := client.Do(ctx, client.B().Ping().Build()).Error(); err != nil {
		return fmt.Errorf("redis ping: %w", err)
	}
	return client.Do(ctx, client.B().Flushdb().Build()).Error()
}
